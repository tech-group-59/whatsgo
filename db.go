package main

import (
	"database/sql"
)

// DBTracker is a struct for tracking messages in a database
type DBTracker struct {
	db *sql.DB
}

func (tracker *DBTracker) Init(config *Config) error {
	// Create table for messages
	_, err := tracker.db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			sender TEXT,
			chat TEXT,
			content TEXT,
			timestamp TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create table for files
	_, err = tracker.db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			path TEXT,
			message_id TEXT,
			FOREIGN KEY(message_id) REFERENCES messages(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// StoreMessage stores a message in the database
func (tracker *DBTracker) StoreMessage(messageID string, sender string, chat string, content string, timestamp string) error {
	_, err := tracker.db.Exec(`INSERT INTO messages (id, sender, chat, content, timestamp) VALUES (?, ?, ?, ?, ?)`,
		messageID, sender, chat, content, timestamp)
	if err != nil {
		log.Errorf("Failed to insert message into database: %v", err)
		return err
	}
	return nil
}

func (tracker *DBTracker) StoreMessageWithFiles(messageID string, sender string, chat string, content string, timestamp string, files []string) error {
	err := tracker.StoreMessage(messageID, sender, chat, content, timestamp)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = tracker.storeFile(messageID, file)
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreFile stores a file in the database
func (tracker *DBTracker) storeFile(messageID string, filePath string) error {
	_, err := tracker.db.Exec(`INSERT INTO files (id, path, message_id) VALUES (?, ?, ?)`,
		messageID, filePath, messageID)
	if err != nil {
		log.Errorf("Failed to insert file into database: %v", err)
		return err
	}
	return nil
}
