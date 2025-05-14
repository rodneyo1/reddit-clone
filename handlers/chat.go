package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	clients    = make(map[string][]*Client)
	broadcast  = make(chan Message)
	register   = make(chan *Client)
	unregister = make(chan *Client)
)

// Message represents a chat message
type Message struct {
	ID          int       `json:"id"`
	TempID       string    `json:"temp_id,omitempty"`
	SenderID    string    `json:"sender_id"`
	RecipientID string    `json:"recipient_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	IsRead      bool      `json:"is_read"`
	SenderUsername string `json:"sender_username"`
    SenderAvatar   string `json:"sender_avatar"`
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
	broadcastUserStatus(userID, true)

	// Handle incoming messages
	go handleIncomingMessages(client)
}

func handleIncomingMessages(client *Client) {
	defer func() {
		unregister <- client
		client.Conn.Close()

		// Update user status to offline
		go func() {
			_, err := db.Exec(`
			UPDATE user_status 
			SET is_online = FALSE, last_seen = ?
			WHERE user_id = ?`,
				time.Now(), client.UserID)
			if err != nil {
				log.Printf("Error updating user status: %v", err)
			}
			broadcastUserStatus(client.UserID, false)
		}()
	}()

	for {
		var msgData map[string]interface{}
        err := client.Conn.ReadJSON(&msgData)
        if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
                log.Printf("Client disconnected: %v", err)
            }
            break
        }

		if msgType, ok := msgData["type"].(string); ok && msgType == "ping" {
            // Send pong response
            client.Conn.WriteJSON(map[string]string{"type": "pong"})
            continue
        }

		if msgData["recipient_id"] == nil || msgData["content"] == nil {
            log.Printf("Invalid message format: %+v", msgData)
            continue
        }

		recipientID, ok1 := msgData["recipient_id"].(string)
        content, ok2 := msgData["content"].(string)
        tempID, _ := msgData["temp_id"].(string) // Optional field

        if !ok1 || !ok2 || recipientID == "" || content == "" {
            log.Printf("Invalid message format: %+v", msgData)
            continue
        }
		 msg := Message{
            SenderID:     client.UserID,
            RecipientID:  recipientID,
            Content:      content,
            CreatedAt:    time.Now(),
            IsRead:       false,
            TempID:       tempID,
        }

		res, err := db.Exec(`
			INSERT INTO messages (sender_id, recipient_id, content, created_at, is_read)
			VALUES (?, ?, ?, ?, ?)`,
			msg.SenderID, msg.RecipientID, msg.Content, msg.CreatedAt, msg.IsRead)
		if err != nil {
			log.Printf("Error saving message: %v", err)
			continue
		}

		

		id, _ := res.LastInsertId()
		msg.ID = int(id)

		var username string
        var avatar sql.NullString
        db.QueryRow("SELECT username, avatar_url FROM users WHERE id = ?", msg.SenderID).Scan(&username, &avatar)
        msg.SenderUsername = username
        if avatar.Valid {
            msg.SenderAvatar = avatar.String
        }

		broadcast <- msg
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
        COALESCE(us.is_online, FALSE) as is_online,
        strftime('%Y-%m-%d %H:%M:%S', COALESCE(us.last_seen, CURRENT_TIMESTAMP)) as last_seen,
        (
            SELECT m.content 
            FROM messages m 
            WHERE (m.sender_id = u.id OR m.recipient_id = u.id) 
            AND (m.sender_id = ? OR m.recipient_id = ?) 
            ORDER BY m.created_at DESC 
            LIMIT 1
        ) as last_message_content,
         COALESCE((
            SELECT m.created_at 
            FROM messages m 
            WHERE (m.sender_id = u.id OR m.recipient_id = u.id) 
            AND (m.sender_id = ? OR m.recipient_id = ?)  
            ORDER BY m.created_at DESC 
            LIMIT 1
        ), CURRENT_TIMESTAMP) as last_message_time,
         COALESCE((
            SELECT COUNT(*) 
            FROM messages m 
            WHERE m.sender_id = u.id 
            AND m.recipient_id = ?  
        ), 0) as unread_count
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

		var (
			lastSeenStr    string
			lastMessageStr string
		)
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&avatarURL,
			&user.IsOnline,
			&lastSeenStr,
			&lastMessage,
			&lastMessageStr,
			&user.UnreadCount,
		)
		user.LastSeen, err = time.Parse("2006-01-02 15:04:05", lastSeenStr)
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

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Transaction begin error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("Transaction rollback error: %v", err)
		}
	}()
	// defer tx.Rollback()

	// Get messages between users using transaction
	rows, err := tx.Query(`
        SELECT 
            m.id,
            m.sender_id,
            m.recipient_id,
            m.content,
            m.created_at,
            m.is_read,
            u.username,
            u.avatar_url
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        WHERE 
            (m.sender_id = ? AND m.recipient_id = ?) OR 
            (m.sender_id = ? AND m.recipient_id = ?)
        ORDER BY m.created_at DESC
        LIMIT 10 OFFSET ?
        `,
		currentUserID, recipientID, recipientID, currentUserID, offset)
	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	var unreadIDs []interface{}

	for rows.Next() {
		var msg Message
		var avatarURL sql.NullString
		var senderUsername string
		// var senderAvatar sql.NullString
		var isRead bool

		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.IsRead,
			&senderUsername,
			&avatarURL,
		)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}

		if avatarURL.Valid {
        msg.SenderAvatar = avatarURL.String
    }

		// Collect unread message IDs
		if msg.RecipientID == currentUserID && !isRead {
			unreadIDs = append(unreadIDs, msg.ID)
		}

		messages = append(messages, msg)
	}

	// Batch update read status if any unread messages
	if len(unreadIDs) > 0 {
		query := "UPDATE messages SET is_read = TRUE WHERE id IN (?" +
			strings.Repeat(",?", len(unreadIDs)-1) + ")"
		_, err = tx.Exec(query, unreadIDs...)
		if err != nil {
			log.Printf("Error marking messages as read: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Commit error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		log.Printf("Error encoding messages: %v", err)
	}
}

// StartChatManager runs the chat manager goroutine
func StartChatManager() {
	for {
		select {
		case client := <-register:
			chatMutex.Lock()
			clients[client.UserID] = append(clients[client.UserID], client)
			chatMutex.Unlock()

		case client := <-unregister:
			chatMutex.Lock()
			if userClients, ok := clients[client.UserID]; ok {
				for i, c := range userClients {
					if c == client {
						clients[client.UserID] = append(userClients[:i], userClients[i+1:]...)
						break
					}
				}
				if len(clients[client.UserID]) == 0 {
					delete(clients, client.UserID)
				}
			}
			chatMutex.Unlock()
		case msg := <-broadcast:
			chatMutex.RLock()
			// Send to sender's clients
			for _, client := range clients[msg.SenderID] {
				if err := client.Conn.WriteJSON(msg); err != nil {
					log.Printf("Write error: %v", err)
					client.Conn.Close()
					unregister <- client
				}
			}
			// Send to recipient's clients
			for _, client := range clients[msg.RecipientID] {
				if err := client.Conn.WriteJSON(msg); err != nil {
					log.Printf("Write error: %v", err)
					client.Conn.Close()
					unregister <- client
				}
			}
			chatMutex.RUnlock()
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
	for _, userClients := range clients {
		for _, client := range userClients {
			err := client.Conn.WriteJSON(statusUpdate)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Printf("Client disconnected unexpectedly: %v", err)
				}
				break
				log.Printf("Error broadcasting status update: %v", err)
				go func(c *Client) {
					unregister <- c
					c.Conn.Close()
				}(client)
			}
		}
	}
}
