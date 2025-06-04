package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
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
		return true // For development only
	},
}

type Client struct {
	UserID  string
	Conn    *websocket.Conn
	writeMu sync.Mutex
}

var (
	clients    = make(map[string][]*Client)
	broadcast  = make(chan Message)
	register   = make(chan *Client)
	unregister = make(chan *Client)
)

func ChatWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

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

	client := &Client{UserID: userID, Conn: conn}
	register <- client

	go handleIncomingMessages(client)
}

func handleIncomingMessages(client *Client) {
	defer func() {
		unregister <- client
		client.Conn.Close()
	}()

	// Improved close handler
	client.Conn.SetCloseHandler(func(code int, text string) error {
		message := websocket.FormatCloseMessage(code, "")
		client.Conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		return nil
	})

	// Better ping/pong handling
	client.Conn.SetPingHandler(func(appData string) error {
		err := client.Conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Timeout() {
			return nil
		}
		return err
	})

	client.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	// Heartbeat ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if err := client.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
				if errors.Is(err, websocket.ErrCloseSent) {
					return
				}
				log.Printf("Ping error: %v", err)
				return
			}
		}
	}()

	for {
		var msgData map[string]interface{}
		if err := client.Conn.ReadJSON(&msgData); err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msgType, ok := msgData["type"].(string); ok && msgType == "ping" {
			client.Conn.WriteJSON(map[string]string{"type": "pong"})
			continue
		}

		if msgType, ok := msgData["type"].(string); ok {
			switch msgType {
			case "typing":
				recipientID, ok := msgData["recipient_id"].(string)
				if !ok {
					continue
				}

				var username string
				err := db.QueryRow("SELECT username FROM users WHERE id = ?", client.UserID).Scan(&username)
				if err != nil {
					continue
				}

				typingStatus := TypingStatus{
					Type:        "typing_status",
					UserID:      client.UserID,
					Username:    username,
					IsTyping:    true,
					RecipientID: recipientID,
				}

				// Send typing status to recipient
				chatMutex.RLock()
				if recipientClients, ok := clients[recipientID]; ok {
					for _, c := range recipientClients {
						c.writeMu.Lock()
						c.Conn.WriteJSON(typingStatus)
						c.writeMu.Unlock()
					}
				}
				chatMutex.RUnlock()

			case "mark_read":
				senderID, ok := msgData["sender_id"].(string)
				if !ok {
					continue
				}

				// Update messages as read in database
				_, err := db.Exec(`
					UPDATE messages 
					SET is_read = TRUE 
					WHERE sender_id = ? AND recipient_id = ? AND is_read = FALSE`,
					senderID, client.UserID)
				if err != nil {
					log.Printf("Error marking messages as read: %v", err)
					continue
				}

				// Notify the sender that their messages were read
				chatMutex.RLock()
				if senderClients, ok := clients[senderID]; ok {
					for _, c := range senderClients {
						c.writeMu.Lock()
						c.Conn.WriteJSON(map[string]interface{}{
							"type":         "messages_read",
							"recipient_id": client.UserID, // Who read the messages
						})
						c.writeMu.Unlock()
					}
				}
				chatMutex.RUnlock()

			case "stop_typing":
				recipientID, ok := msgData["recipient_id"].(string)
				if !ok {
					continue
				}

				typingStatus := TypingStatus{
					Type:        "typing_status",
					UserID:      client.UserID,
					IsTyping:    false,
					RecipientID: recipientID,
				}

				chatMutex.RLock()
				if recipientClients, ok := clients[recipientID]; ok {
					for _, c := range recipientClients {
						c.writeMu.Lock()
						c.Conn.WriteJSON(typingStatus)
						c.writeMu.Unlock()
					}
				}
				chatMutex.RUnlock()
			}
		}

		recipientID, ok1 := msgData["recipient_id"].(string)
		content, ok2 := msgData["content"].(string)
		if !ok1 || !ok2 || recipientID == "" || content == "" {
			continue
		}

		msg := Message{
			SenderID:    client.UserID,
			RecipientID: recipientID,
			Content:     content,
			CreatedAt:   time.Now(),
			TempID:      msgData["temp_id"].(string),
		}

		res, err := db.Exec(`
			INSERT INTO messages (sender_id, recipient_id, content, created_at, is_read)
			VALUES (?, ?, ?, ?, ?)`,
			msg.SenderID, msg.RecipientID, msg.Content, msg.CreatedAt, false)
		if err != nil {
			log.Printf("Message save error: %v", err)
			continue
		}

		id, _ := res.LastInsertId()
		msg.ID = int(id)

		var username string
		var avatar sql.NullString
		if err := db.QueryRow("SELECT username, avatar_url FROM users WHERE id = ?", msg.SenderID).
			Scan(&username, &avatar); err == nil {
			msg.SenderUsername = username
			if avatar.Valid {
				msg.SenderAvatar = avatar.String
			}
		}

		broadcast <- msg
	}
}

func ChatUsersHandler(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var currentUserID string
	if err := db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).
		Scan(&currentUserID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.Query(`
		SELECT 
			u.id, 
			u.username, 
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(us.is_online, FALSE) as is_online,
			datetime(COALESCE(us.last_seen, CURRENT_TIMESTAMP)) as last_seen,
			COALESCE((
				SELECT m.content FROM messages m 
				WHERE (m.sender_id = u.id OR m.recipient_id = u.id) 
				AND (m.sender_id = ? OR m.recipient_id = ?) 
				ORDER BY m.created_at DESC LIMIT 1
			), '') as last_message,
			COALESCE((
				SELECT datetime(m.created_at) FROM messages m 
				WHERE (m.sender_id = u.id OR m.recipient_id = u.id) 
				AND (m.sender_id = ? OR m.recipient_id = ?)  
				ORDER BY m.created_at DESC LIMIT 1
			), datetime('now')) as last_message_time,
			COALESCE((
				SELECT COUNT(*) FROM messages m 
				WHERE m.sender_id = u.id 
				AND m.recipient_id = ? 
				AND m.is_read = FALSE
			), 0) as unread_count
		FROM users u
		LEFT JOIN user_status us ON u.id = us.user_id
		WHERE u.id != ?`,
		currentUserID, currentUserID,
		currentUserID, currentUserID,
		currentUserID,
		currentUserID)
	if err != nil {
		log.Printf("User query error: %v", err)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	defer rows.Close()

	type User struct {
		ID              string    `json:"id"`
		Username        string    `json:"username"`
		AvatarURL       string    `json:"avatar_url"`
		IsOnline        bool      `json:"is_online"`
		LastSeen        time.Time `json:"last_seen"`
		LastMessage     string    `json:"last_message"`
		LastMessageTime time.Time `json:"last_message_time"`
		UnreadCount     int       `json:"unread_count"`
	}

	var users []User
	for rows.Next() {
		var u User
		var lastSeen, lastMsgTime string

		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.AvatarURL,
			&u.IsOnline,
			&lastSeen,
			&u.LastMessage,
			&lastMsgTime,
			&u.UnreadCount,
		); err != nil {
			continue
		}

		u.LastSeen, _ = time.Parse("2006-01-02 15:04:05", lastSeen)
		u.LastMessageTime, _ = time.Parse("2006-01-02 15:04:05", lastMsgTime)

		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func ChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var currentUserID string
	if err := db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).
		Scan(&currentUserID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	recipientID := r.URL.Query().Get("recipient_id")
	if recipientID == "" {
		http.Error(w, "Recipient required", http.StatusBadRequest)
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// Default to 10 messages if no limit specified
	if limit <= 0 {
		limit = 10
	}
	// Cap maximum limit to prevent excessive loads
	if limit > 50 {
		limit = 50
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT 
			m.id,
			m.sender_id,
			m.recipient_id,
			m.content,
			m.created_at,
			m.is_read,
			u.username,
			COALESCE(u.avatar_url, ''),
			CASE WHEN m.sender_id = ? THEN 1 ELSE 0 END
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE (m.sender_id = ? AND m.recipient_id = ?)
			OR (m.sender_id = ? AND m.recipient_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?`,
		currentUserID,
		currentUserID, recipientID,
		recipientID, currentUserID,
		limit, offset)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	var unreadIDs []interface{}

	for rows.Next() {
		var msg Message
		var avatar string
		var isOwner int

		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.IsRead,
			&msg.SenderUsername,
			&avatar,
			&isOwner,
		); err != nil {
			continue
		}

		msg.IsOwner = isOwner == 1
		msg.SenderAvatar = avatar

		if msg.RecipientID == currentUserID && !msg.IsRead {
			unreadIDs = append(unreadIDs, msg.ID)
		}
		messages = append(messages, msg)
	}

	if len(unreadIDs) > 0 {
		query := "UPDATE messages SET is_read = TRUE WHERE id IN (?" +
			strings.Repeat(",?", len(unreadIDs)-1) + ")"
		tx.Exec(query, unreadIDs...)
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(messages)
}

func StartChatManager() {
	for {
		select {
		case client := <-register:
			chatMutex.Lock()
			clients[client.UserID] = append(clients[client.UserID], client)

			// Only update status if first connection
			if len(clients[client.UserID]) == 1 {
				updateUserStatus(client.UserID, true)
				go broadcastUserStatusToAll(client.UserID, true)
			}
			chatMutex.Unlock()

		case client := <-unregister:
			chatMutex.Lock()
			if userClients, ok := clients[client.UserID]; ok {
				// Find and remove client
				for i, c := range userClients {
					if c == client {
						clients[client.UserID] = append(userClients[:i], userClients[i+1:]...)
						break
					}
				}

				// Update status only when last connection closes
				if len(clients[client.UserID]) == 0 {
					delete(clients, client.UserID)
					updateUserStatus(client.UserID, false)
					go broadcastUserStatusToAll(client.UserID, false)
				}
			}
			chatMutex.Unlock()

		case msg := <-broadcast:
			chatMutex.RLock()
			// Send to sender
			if senderClients, ok := clients[msg.SenderID]; ok {
				for _, c := range senderClients {
					go func(client *Client) {
						client.writeMu.Lock()
						defer client.writeMu.Unlock()
						sendMsg := msg
						sendMsg.IsOwner = true
						client.Conn.WriteJSON(sendMsg)
					}(c)
				}
			}

			// Send to recipient
			if recipientClients, ok := clients[msg.RecipientID]; ok {
				for _, c := range recipientClients {
					go func(client *Client) {
						client.writeMu.Lock()
						defer client.writeMu.Unlock()
						sendMsg := msg
						sendMsg.IsOwner = false
						client.Conn.WriteJSON(sendMsg)
					}(c)
				}
			}
			chatMutex.RUnlock()
		}
	}
}

// Improved status broadcasting
func broadcastUserStatusToAll(userID string, isOnline bool) {
	chatMutex.RLock()
	defer chatMutex.RUnlock()

	// Get all online users
	onlineUsers := make(map[string]bool)
	for uid := range clients {
		onlineUsers[uid] = true
	}

	// Broadcast to all connected clients
	for uid, userClients := range clients {
		// Only send to users who are not the one changing status
		if uid == userID {
			continue
		}

		for _, client := range userClients {
			status := map[string]interface{}{
				"type":      "status_update",
				"user_id":   userID,
				"is_online": isOnline,
				"timestamp": time.Now().Unix(),
			}
			go func(c *Client) {
				c.writeMu.Lock()
				defer c.writeMu.Unlock()
				c.Conn.WriteJSON(status)
			}(client)
		}
	}
}

// Simplified status update
func updateUserStatus(userID string, online bool) {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO user_status 
		(user_id, is_online, last_seen)
		VALUES (?, ?, ?)`,
		userID, online, time.Now())
	if err != nil {
		log.Printf("Status update error: %v", err)
	}
}
