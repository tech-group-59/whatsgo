package main

import (
	"database/sql"
)

//@TODO: Add support of postgresql if needed

// Create tables for messages and media files if they don't exist
func UpgradeDb(db *sql.DB, err error) {
	if err != nil {
		log.Errorf("Failed to connect to database: %v", err)
		return
	}

	// Create table for messages
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS messages (
            id TEXT PRIMARY KEY,
            sender TEXT,
            chat TEXT,
            content TEXT,
            timestamp TEXT
        )
    `)
	if err != nil {
		log.Errorf("Failed to create messages table: %v", err)
		return
	}

	// Create table for files
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS files (
            id TEXT PRIMARY KEY,
            path TEXT,
            message_id TEXT,
            FOREIGN KEY(message_id) REFERENCES messages(id)
        )
    `)
	if err != nil {
		log.Errorf("Failed to create files table: %v", err)
		return
	}
}

// StoreMessage stores a message in the database
func StoreMessage(db *sql.DB, messageID string, sender string, chat string, content string, timestamp string) error {
	_, err := db.Exec(`INSERT INTO messages (id, sender, chat, content, timestamp) VALUES (?, ?, ?, ?, ?)`,
		messageID, sender, chat, content, timestamp)
	if err != nil {
		log.Errorf("Failed to insert message into database: %v", err)
		return err
	}
	return nil
}

// StoreFile stores a file in the database
func StoreFile(db *sql.DB, messageID string, filePath string) error {
	_, err := db.Exec(`INSERT INTO files (id, path, message_id) VALUES (?, ?, ?)`,
		messageID, filePath, messageID)
	if err != nil {
		log.Errorf("Failed to insert file into database: %v", err)
		return err
	}
	return nil
}
