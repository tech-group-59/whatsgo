package main

import "database/sql"

type Tracker interface {
	Init(config *Config) error
	StoreMessage(messageID string, sender string, chat string, content string, timestamp string) error
	StoreMessageWithFiles(messageID string, sender string, chat string, content string, timestamp string, files []string) error
}

func CreateTrackers(config *Config, db *sql.DB) []Tracker {
	var trackers []Tracker
	trackers = append(trackers, &DBTracker{db: db})
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

func processMessageByTracker(tracker Tracker, messageID string, sender string, chat string, content string, timestamp string, files []string) error {
	var err error

	if len(files) > 0 {
		err = tracker.StoreMessageWithFiles(messageID, sender, chat, content, timestamp, files)
	} else {
		err = tracker.StoreMessage(messageID, sender, chat, content, timestamp)
	}

	return err
}

func ProcessMessage(trackers []Tracker, messageID string, sender string, chat string, content string, timestamp string, files []string) {
	for _, tracker := range trackers {
		err := processMessageByTracker(tracker, messageID, sender, chat, content, timestamp, files)
		if err != nil {
			log.Errorf("Failed to store message in tracker: %v", err)
		}
	}
}
