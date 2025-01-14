package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"sync"
	"time"
)

const FileWebPathPrefix = "/files"

type Server struct {
	DB              *sql.DB
	config          *Config
	fileStoragePath string
	clients         map[*websocket.Conn]bool
	broadcast       chan []byte
	upgrader        websocket.Upgrader
	mu              sync.Mutex
}

func (s *Server) InitWebSocket() {
	s.clients = make(map[*websocket.Conn]bool)
	s.broadcast = make(chan []byte)
	s.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins
			return true
		},
	}
}

// WebSocket endpoint
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			break
		}
	}
}

// Broadcast messages to all connected clients
func (s *Server) broadcastToClients(message TrackableMessage) {
	//create WebMessage
	webMsg := WebMessage{
		ID:        message.MessageID,
		Sender:    message.Sender,
		Chat:      message.Chat,
		Content:   message.Content,
		Timestamp: message.Timestamp,
		Filename:  nil,
	}
	if len(message.Files) > 0 {
		filename := FileWebPathPrefix + strings.TrimPrefix(message.Files[0], s.fileStoragePath)
		webMsg.Filename = &filename
	}

	//convert the message to JSON
	wsMsg, err := json.Marshal(webMsg)
	if err != nil {
		return // Ignore messages that can't be marshalled
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, wsMsg)
		if err != nil {
			client.Close()
			delete(s.clients, client)
		}
	}
}

func (s *Server) getDBChatsHandler(w http.ResponseWriter, r *http.Request) {
	// Return chats from config
	json.NewEncoder(w).Encode(s.config.Chats)
}

type WebMessage struct {
	ID        string  `json:"id"`
	Sender    string  `json:"sender"`
	Chat      string  `json:"chat"`
	Content   string  `json:"content"`
	Timestamp string  `json:"timestamp"`
	Filename  *string `json:"filename"`
}

func removeFilePrefixFromWebMessage(message *WebMessage, prefix string) {
	if message.Filename != nil {
		*message.Filename = FileWebPathPrefix + strings.TrimPrefix(*message.Filename, prefix)
	}
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
        SELECT messages.id, sender, chat, content, timestamp, path
        FROM messages
        LEFT JOIN files ON messages.id = files.message_id
        WHERE date(substr(timestamp,0,11)) >= date(?) AND date(substr(timestamp,0,11)) <= date(?)`

	// If a content filter is provided, add it to the query
	var args []interface{}
	args = append(args, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))

	if content != "" {
		sqlQuery += " AND (content LIKE ? OR lower(content) LIKE ?)"
		loweredContent := "%" + strings.ToLower(content) + "%"
		args = append(args, "%"+content+"%", loweredContent)
	}

	// Order by timestamp in descending order
	sqlQuery += " ORDER BY timestamp DESC"

	// Query messages from the DB
	messages, err := s.DB.Query(sqlQuery, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get messages: %v", err), http.StatusInternalServerError)
		return
	}
	defer messages.Close()

	var messageList []WebMessage
	for messages.Next() {
		var message WebMessage
		if err := messages.Scan(&message.ID, &message.Sender, &message.Chat, &message.Content, &message.Timestamp, &message.Filename); err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan message: %v", err), http.StatusInternalServerError)
			return
		}
		removeFilePrefixFromWebMessage(&message, s.fileStoragePath)
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

func CreateServer(config *Config, db *sql.DB) *Server {
	server := &Server{
		DB:              db,
		config:          config,
		fileStoragePath: config.FileStoragePath,
	}
	server.InitWebSocket() // Initialize WebSocket
	return server
}

func RunServer(server *Server) {
	// Create a new ServeMux (multiplexer) for your routes
	mux := http.NewServeMux()

	// Register your handlers
	mux.HandleFunc("/chats", server.getDBChatsHandler)
	mux.HandleFunc("/messages", server.getDBMessagesHandler)
	mux.HandleFunc("/ws", server.handleWebSocket) // WebSocket endpoint

	// Serve static files from the "data" directory at the "files" path
	fsData := http.StripPrefix(FileWebPathPrefix+"/", http.FileServer(http.Dir(server.fileStoragePath)))
	mux.Handle(FileWebPathPrefix+"/", fsData)

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
