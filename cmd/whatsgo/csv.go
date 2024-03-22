package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

type CSVTracker struct {
	csvWriters map[string]*csv.Writer
	config     CSVConfig
}

func (tracker *CSVTracker) Init(config *Config) error {
	tracker.csvWriters = make(map[string]*csv.Writer)
	tracker.config = config.CSV
	return nil
}

func (tracker *CSVTracker) TrackMessage(message *TrackableMessage) error {
	csvWriter, exists := tracker.csvWriters[message.Chat]
	if !exists {
		fileName := fmt.Sprintf("%s/%s.csv", tracker.config.Path, message.Chat)
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Errorf("Failed to open CSV file: %v", err)
			return err
		}
		csvWriter = csv.NewWriter(file)
		tracker.csvWriters[message.Chat] = csvWriter
	}

	record := []string{
		message.MessageID,
		message.Sender,
		message.Chat,
		message.Content,
		message.ParsedContent,
		message.Timestamp,
	}

	for _, file := range message.Files {
		record = append(record, file)
	}

	if err := csvWriter.Write(record); err != nil {
		log.Errorf("Failed to write record to CSV: %v", err)
		return err
	}

	csvWriter.Flush()
	return csvWriter.Error()
}
