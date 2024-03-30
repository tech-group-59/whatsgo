package main

import (
	"database/sql"
	"time"
)

type MessageMetadata struct {
	Date      string
	Folder    string
	Timestamp time.Time
}

type TrackableMessage struct {
	MessageID     string
	Sender        string
	Chat          string
	Content       string
	ParsedContent string
	Timestamp     string
	Files         []string
	Metadata      MessageMetadata
}

type Tracker interface {
	Init(config *Config) error
	TrackMessage(message *TrackableMessage) error
}

func CreateTrackers(config *Config, db *sql.DB) []Tracker {
	var trackers []Tracker
	if config.OCR.Enabled {
		trackers = append(trackers, &OCRTracker{})
	}

	trackers = append(trackers, &DBTracker{db: db})
	if config.CSV.Enabled {
		trackers = append(trackers, &CSVTracker{})
	}
	if config.GoogleCloud.Enabled {
		trackers = append(trackers, &CloudTracker{})
	}

	// Init all trackers
	for _, tracker := range trackers {
		err := tracker.Init(config)
		if err != nil {
			log.Errorf("Failed to initialize tracker: %v", err)
		}

	}
	return trackers
}

func ProcessMessage(trackers []Tracker, messageID string, sender string, chat string, content string, timestamp string, files []string, metadata MessageMetadata) error {
	message := TrackableMessage{
		MessageID:     messageID,
		Sender:        sender,
		Chat:          chat,
		Content:       content,
		ParsedContent: "",
		Timestamp:     timestamp,
		Files:         files,
		Metadata:      metadata,
	}

	for _, tracker := range trackers {
		err := tracker.TrackMessage(&message)
		if err != nil {
			log.Errorf("Failed to store message in tracker: %v", err)
		}
	}
	return nil
}
