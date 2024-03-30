package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

type CSVTracker struct {
	config CSVConfig
	chats  []Chat
}

func (tracker *CSVTracker) Init(config *Config) error {
	tracker.config = config.CSV
	tracker.chats = config.Chats
	return nil
}

func (tracker *CSVTracker) TrackMessage(message *TrackableMessage) error {
	// Create the directory path
	dirPath := fmt.Sprintf("%s/%s/%s", tracker.config.Path, message.Metadata.Folder, message.Metadata.Date)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0755)
	}

	// Open the CSV file
	fileName := fmt.Sprintf("%s/messages.csv", dirPath)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		log.Errorf("Failed to open CSV file: %v", err)
		return err
	}

	// Write the message to the CSV file
	csvWriter := csv.NewWriter(file)
	record := []string{
		message.MessageID,
		message.Sender,
		message.Chat,
		strings.ReplaceAll(message.Content, "\n", " "),
		strings.ReplaceAll(message.ParsedContent, "\n", " "),
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
