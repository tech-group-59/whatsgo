package main

import (
	"database/sql"
	"time"
)

// DBTracker is a struct for tracking messages in a database
type DBTracker struct {
	db     *sql.DB
	config *Config
}

func (tracker *DBTracker) Init(config *Config) error {
	tracker.config = config
	// Create table for messages
	_, err := tracker.db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			sender TEXT,
			chat TEXT,
			content TEXT,
			parsed_content TEXT,
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

	// Check if parsedContent column exists and create it if not
	_, err = tracker.db.Exec(`SELECT parsed_content FROM messages LIMIT 1`)
	if err != nil {
		_, err = tracker.db.Exec(`ALTER TABLE messages ADD COLUMN parsed_content TEXT DEFAULT ''`)
		if err != nil {
			return err
		}
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

func (tracker *DBTracker) GetMessagesByChat(chat string, date time.Time) ([]TrackableMessage, error) {
	var rows *sql.Rows
	var err error
	query := `SELECT id, sender, chat, content, parsed_content, timestamp FROM messages WHERE chat = ? AND timestamp = date(?)`
	rows, err = tracker.db.Query(query, chat, date.Format("2006-01-02"))

	if err != nil {
		log.Errorf("Failed to query messages from database: %v", err)
		return nil, err
	}
	defer rows.Close()

	folder := chat
	for _, c := range tracker.config.Chats {
		if c.ID == chat && c.Alias != "" {
			folder = c.Alias
			break
		}
	}

	var messages []TrackableMessage
	for rows.Next() {
		var message TrackableMessage
		err := rows.Scan(&message.MessageID, &message.Sender, &message.Chat, &message.Content, &message.ParsedContent, &message.Timestamp)
		if err != nil {
			log.Errorf("Failed to scan message from database: %v", err)
			return nil, err
		}
		// parse timestamp
		message.Metadata.Timestamp, err = time.Parse("2006-01-02 15:04:05 -0700 MST", message.Timestamp)
		message.Metadata.Folder = folder
		message.Metadata.Date = message.Metadata.Timestamp.Format("02.01.2006")

		files, err := tracker.GetFilesByMessage(message.MessageID)
		if err != nil {
			return nil, err
		}
		message.Files = files

		messages = append(messages, message)
	}

	return messages, nil
}

func (tracker *DBTracker) GetFilesByMessage(messageID string) ([]string, error) {
	rows, err := tracker.db.Query(`SELECT path FROM files WHERE message_id = ?`, messageID)
	if err != nil {
		log.Errorf("Failed to query files from database: %v", err)
		return nil, err
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var file string
		err := rows.Scan(&file)
		if err != nil {
			log.Errorf("Failed to scan file from database: %v", err)
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// StoreMessage stores a message in the database
func (tracker *DBTracker) storeMessage(messageID string, sender string, chat string, content string, parsedContent string, timestamp string) error {
	_, err := tracker.db.Exec(`INSERT INTO messages (id, sender, chat, content, parsed_content, timestamp) VALUES (?, ?, ?, ?, ?, ?)`,
		messageID, sender, chat, content, parsedContent, timestamp)
	if err != nil {
		log.Errorf("Failed to insert message into database: %v", err)
		return err
	}
	return nil
}

func (tracker *DBTracker) TrackMessage(message *TrackableMessage) error {

	err := tracker.storeMessage(message.MessageID, message.Sender, message.Chat, message.Content, message.ParsedContent, message.Timestamp)
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
