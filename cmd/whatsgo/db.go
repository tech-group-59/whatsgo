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

func (tracker *DBTracker) GetChats() ([]string, error) {
	rows, err := tracker.db.Query(`SELECT DISTINCT chat FROM messages`)
	if err != nil {
		log.Errorf("Failed to query messages from database: %v", err)
		return nil, err
	}
	defer rows.Close()

	var chats []string
	for rows.Next() {
		var chat string
		err := rows.Scan(&chat)
		if err != nil {
			log.Errorf("Failed to scan message from database: %v", err)
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

func (tracker *DBTracker) GetMessagesByChat(chat string) ([]TrackableMessage, error) {
	rows, err := tracker.db.Query(`SELECT id, sender, chat, content, timestamp FROM messages WHERE chat = ?`, chat)
	if err != nil {
		log.Errorf("Failed to query messages from database: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []TrackableMessage
	for rows.Next() {
		var message TrackableMessage
		err := rows.Scan(&message.MessageID, &message.Sender, &message.Chat, &message.Content, &message.Timestamp)
		if err != nil {
			log.Errorf("Failed to scan message from database: %v", err)
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// StoreMessage stores a message in the database
func (tracker *DBTracker) storeMessage(messageID string, sender string, chat string, content string, timestamp string) error {
	_, err := tracker.db.Exec(`INSERT INTO messages (id, sender, chat, content, timestamp) VALUES (?, ?, ?, ?, ?)`,
		messageID, sender, chat, content, timestamp)
	if err != nil {
		log.Errorf("Failed to insert message into database: %v", err)
		return err
	}
	return nil
}

func (tracker *DBTracker) TrackMessage(message *TrackableMessage) error {

	err := tracker.storeMessage(message.MessageID, message.Sender, message.Chat, message.Content, message.Timestamp)
	if err != nil {
		return err
	}

	for _, file := range message.Files {
		err = tracker.storeFile(message.MessageID, file)
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
