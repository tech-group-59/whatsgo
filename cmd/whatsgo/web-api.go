package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	DB     *sql.DB
	config *Config
}

func (s *Server) getDBChatsHandler(w http.ResponseWriter, r *http.Request) {
	// Return chats from config
	json.NewEncoder(w).Encode(s.config.Chats)
}

type Message struct {
	ID            string `json:"id"`
	Sender        string `json:"sender"`
	Chat          string `json:"chat"`
	Content       string `json:"content"`
	Timestamp     string `json:"timestamp"`
	ParsedContent string `json:"parsed_content"`
}

func (s *Server) getDBMessagesHandler(w http.ResponseWriter, r *http.Request) {
	dateFromStr := r.URL.Query().Get("from")
	dateToStr := r.URL.Query().Get("to")
	content := r.URL.Query().Get("content")

	if dateFromStr == "" || dateToStr == "" {
		http.Error(w, "Missing date_from or date_to", http.StatusBadRequest)
		return
	}

	dateFrom, err := time.Parse("02.01.2006", dateFromStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid date_from format: %v", err), http.StatusBadRequest)
		return
	}

	dateTo, err := time.Parse("02.01.2006", dateToStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid date_to format: %v", err), http.StatusBadRequest)
		return
	}

	// Prepare the base SQL query
	sqlQuery := `
        SELECT id, sender, chat, content, timestamp, parsed_content
        FROM messages
        WHERE date(substr(timestamp,0,11)) >= date(?) AND date(substr(timestamp,0,11)) <= date(?)`

	// If a content filter is provided, add it to the query
	var args []interface{}
	args = append(args, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))

	if content != "" {
		sqlQuery += " AND (lower(content) LIKE ? OR lower(parsed_content) LIKE ?)"
		loweredContent := "%" + strings.ToLower(content) + "%"
		args = append(args, loweredContent, loweredContent)
	}

	// Query messages from the DB
	messages, err := s.DB.Query(sqlQuery, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get messages: %v", err), http.StatusInternalServerError)
		return
	}
	defer messages.Close()

	var messageList []Message
	for messages.Next() {
		var message Message
		if err := messages.Scan(&message.ID, &message.Sender, &message.Chat, &message.Content, &message.Timestamp, &message.ParsedContent); err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan message: %v", err), http.StatusInternalServerError)
			return
		}
		messageList = append(messageList, message)
	}

	json.NewEncoder(w).Encode(messageList)
}

// Middleware to handle CORS
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins (you can restrict this to specific origins)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Specify the allowed methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Specify the allowed headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Allow credentials (optional, use if needed)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func RunServer(config *Config, db *sql.DB) {
	server := &Server{
		DB:     db,
		config: config,
	}

	// Create a new ServeMux (multiplexer) for your routes
	mux := http.NewServeMux()

	// Register your handlers
	mux.HandleFunc("/chats", server.getDBChatsHandler)
	mux.HandleFunc("/messages", server.getDBMessagesHandler)

	// Serve static files from the "./static" directory at the root path "/"
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	// Apply CORS middleware to the mux
	handlerWithCORS := server.corsMiddleware(mux)

	httpPort := "8080"
	log.Infof("Starting HTTP server on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, handlerWithCORS); err != nil {
		log.Errorf("Failed to start HTTP server: %v", err)
	}
}
