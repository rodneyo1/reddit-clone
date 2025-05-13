package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var chatMutex sync.RWMutex

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development only, restrict in production
	},
}

// WebSocket connection structure
type Client struct {
	UserID string
	Conn   *websocket.Conn
}

// Global variables for chat management
var (
	clients    = make(map[string]*Client)
	broadcast  = make(chan Message)
	register   = make(chan *Client)
	unregister = make(chan *Client)
)

// Message represents a chat message
type Message struct {
	ID          int       `json:"id"`
	SenderID    string    `json:"sender_id"`
	RecipientID string    `json:"recipient_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	IsRead      bool      `json:"is_read"`
}

// ChatWebsocketHandler handles WebSocket connections for chat
func ChatWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get user ID from session
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		conn.Close()
		return
	}

	var userID string
	err = db.QueryRow(`
    SELECT user_id FROM sessions 
    WHERE session_id = ? AND expires_at > ?`,
		sessionCookie.Value, time.Now()).Scan(&userID)
	if err != nil {
		conn.Close()
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
	}

	register <- client

	// Update user status to online
	_, err = db.Exec(`
		INSERT OR REPLACE INTO user_status (user_id, is_online, last_seen)
		VALUES (?, TRUE, ?)`,
		userID, time.Now())
	if err != nil {
		log.Printf("Error updating user status: %v", err)
	}

	// Handle incoming messages
	go handleIncomingMessages(client)
}

func handleIncomingMessages(client *Client) {
	defer func() {
		unregister <- client
		client.Conn.Close()

		// Update user status to offline
		_, err := db.Exec(`
			UPDATE user_status 
			SET is_online = FALSE, last_seen = ?
			WHERE user_id = ?`,
			time.Now(), client.UserID)
		if err != nil {
			log.Printf("Error updating user status: %v", err)
		}
	}()

	for {
		_, msgBytes, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var message Message
		if err := json.Unmarshal(msgBytes, &message); err != nil {
			log.Printf("Error decoding message: %v", err)
			continue
		}

		message.SenderID = client.UserID
		message.CreatedAt = time.Now()
		message.IsRead = false

		// Save message to database
		_, err = db.Exec(`
			INSERT INTO messages (sender_id, recipient_id, content, created_at, is_read)
			VALUES (?, ?, ?, ?, ?)`,
			message.SenderID, message.RecipientID, message.Content, message.CreatedAt, message.IsRead)
		if err != nil {
			log.Printf("Error saving message: %v", err)
			continue
		}

		broadcast <- message
	}
}

// ChatUsersHandler returns the list of users for the chat interface
func ChatUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from session
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var currentUserID string
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&currentUserID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all users with their status and last message info
	rows, err := db.Query(`
		SELECT 
        u.id, 
        u.username, 
        u.avatar_url,
        us.is_online,
        us.last_seen,
        (
            SELECT m.content 
            FROM messages m 
            WHERE (m.sender_id = u.id OR m.recipient_id = u.id) 
            AND (m.sender_id = ? OR m.recipient_id = ?) 
            ORDER BY m.created_at DESC 
            LIMIT 1
        ) as last_message_content,
        (
            SELECT m.created_at 
            FROM messages m 
            WHERE (m.sender_id = u.id OR m.recipient_id = u.id) 
            AND (m.sender_id = ? OR m.recipient_id = ?)  
            ORDER BY m.created_at DESC 
            LIMIT 1
        ) as last_message_time,
        (
            SELECT COUNT(*) 
            FROM messages m 
            WHERE m.sender_id = u.id 
            AND m.recipient_id = ?  
        ) as unread_count
    FROM users u
    LEFT JOIN user_status us ON u.id = us.user_id
    WHERE u.id != ?`, // Needs 1 param
		currentUserID, currentUserID, // For first subquery
		currentUserID, currentUserID, // For second subquery
		currentUserID, // For unread_count
		currentUserID)
	if err != nil {
		log.Printf("Database query error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()

	type UserWithStatus struct {
		ID              string    `json:"id"`
		Username        string    `json:"username"`
		AvatarURL       string    `json:"avatar_url"`
		IsOnline        bool      `json:"is_online"`
		LastSeen        time.Time `json:"last_seen"`
		LastMessage     string    `json:"last_message"`
		LastMessageTime time.Time `json:"last_message_time"`
		UnreadCount     int       `json:"unread_count"`
	}

	var users []UserWithStatus
	for rows.Next() {
		var user UserWithStatus
		var lastMessage sql.NullString
		var lastMessageTime sql.NullTime
		var avatarURL sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&avatarURL,
			&user.IsOnline,
			&user.LastSeen,
			&lastMessage,
			&lastMessageTime,
			&user.UnreadCount,
		)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}

		if avatarURL.Valid {
			user.AvatarURL = avatarURL.String
		}
		if lastMessage.Valid {
			user.LastMessage = lastMessage.String
		}
		if lastMessageTime.Valid {
			user.LastMessageTime = lastMessageTime.Time
		}

		users = append(users, user)
	}
	if users == nil {
		users = []UserWithStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Error encoding users: %v", err)
		// Fallback to empty array
		json.NewEncoder(w).Encode([]interface{}{})
	}
}

// ChatMessagesHandler returns the message history between two users
func ChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from session
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var currentUserID string
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&currentUserID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get recipient ID from query params
	recipientID := r.URL.Query().Get("recipient_id")
	if recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Get offset for pagination
	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
	}

	// Get messages between users
	rows, err := db.Query(`
		SELECT 
			m.id,
			m.sender_id,
			m.recipient_id,
			m.content,
			m.created_at,
			m.is_read,
			u.username as sender_username,
			u.avatar_url as sender_avatar
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE 
			(m.sender_id = ? AND m.recipient_id = ?) OR 
			(m.sender_id = ? AND m.recipient_id = ?)
		ORDER BY m.created_at DESC
		LIMIT 10 OFFSET ?`,
		currentUserID, recipientID, recipientID, currentUserID, offset)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var senderUsername string
		var senderAvatar sql.NullString
		var isRead bool

		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.CreatedAt,
			&isRead,
			&senderUsername,
			&senderAvatar,
		)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}

		// Mark messages as read if they're sent to the current user
		if msg.RecipientID == currentUserID && !isRead {
			_, err = db.Exec("UPDATE messages SET is_read = TRUE WHERE id = ?", msg.ID)
			if err != nil {
				log.Printf("Error marking message as read: %v", err)
			}
		}

		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// StartChatManager runs the chat manager goroutine
func StartChatManager() {
	for {
		select {
		case client := <-register:
			chatMutex.Lock()
			clients[client.UserID] = client
			chatMutex.Unlock()

		case client := <-unregister:
			chatMutex.Lock()
			delete(clients, client.UserID)
			chatMutex.Unlock()
		}
	}
}

func broadcastUserStatus(userID string, isOnline bool) {
	// Get username for the status update
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		log.Printf("Error getting username for status broadcast: %v", err)
		return
	}

	statusUpdate := map[string]interface{}{
		"type":      "status_update",
		"user_id":   userID,
		"username":  username,
		"is_online": isOnline,
		"timestamp": time.Now(),
	}

	// Broadcast to all connected clients
	for _, client := range clients {
		err := client.Conn.WriteJSON(statusUpdate)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Client disconnected unexpectedly: %v", err)
			}
			log.Printf("Error broadcasting status update: %v", err)
			client.Conn.Close()
			delete(clients, client.UserID)
		}
	}
}
