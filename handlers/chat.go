package handlers

import (
	"database/sql"
	"encoding/json"
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
		return true // For development only
	},
}

type Client struct {
	UserID string
	Conn   *websocket.Conn
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
	

	// Close existing connections
	chatMutex.Lock()
	if existing, ok := clients[userID]; ok {
		for _, c := range existing {
			c.Conn.Close()
		}
	}
	chatMutex.Unlock()

	client := &Client{UserID: userID, Conn: conn}
	register <- client

	_, err = db.Exec(`
		INSERT OR REPLACE INTO user_status (user_id, is_online, last_seen)
		VALUES (?, TRUE, ?)`,
		userID, time.Now())
	if err != nil {
		log.Printf("Status update error: %v", err)
	}

	go handleIncomingMessages(client)
}

func handleIncomingMessages(client *Client) {
	defer func() {
		unregister <- client
		client.Conn.Close()
		updateUserStatus(client.UserID, false)
	}()


	client.Conn.SetCloseHandler(func(code int, text string) error {
        log.Printf("Connection closing: %d %s", code, text)
        return nil
    })

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		var msgData map[string]interface{}
		if err := client.Conn.ReadJSON(&msgData); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msgType, ok := msgData["type"].(string); ok && msgType == "ping" {
			client.Conn.WriteJSON(map[string]string{"type": "pong"})
			continue
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
				WHERE m.sender_id = u.id AND m.recipient_id = ?
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

	offset, _ := fmt.Sscanf(r.URL.Query().Get("offset"), "%d")

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
		LIMIT 10 OFFSET ?`,
		currentUserID,
		currentUserID, recipientID,
		recipientID, currentUserID,
		offset)
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
			if existing, ok := clients[client.UserID]; ok {
				for _, c := range existing {
					c.Conn.Close()
				}
			}
			clients[client.UserID] = []*Client{client}
			chatMutex.Unlock()
			broadcastUserStatus(client.UserID, true)

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
					go broadcastUserStatus(client.UserID, false)
				}
			}
			chatMutex.Unlock()

		case msg := <-broadcast:
			chatMutex.RLock()
			defer chatMutex.RUnlock()

			if senderClients, ok := clients[msg.SenderID]; ok {
				for _, client := range senderClients {
					senderMsg := msg
					senderMsg.IsOwner = true
					client.Conn.WriteJSON(senderMsg)
				}
			}

			if recipientClients, ok := clients[msg.RecipientID]; ok {
				for _, client := range recipientClients {
					senderMsg := msg
					senderMsg.IsOwner = false
					client.Conn.WriteJSON(senderMsg)
				}
			}
		}
	}
}

func broadcastUserStatus(userID string, isOnline bool) {
	var lastSeen time.Time
    var username string
    err := db.QueryRow(`
        SELECT u.username, us.last_seen 
        FROM user_status us
		JOIN users u ON us.user_id = u.id 
        WHERE us.user_id = ?`, userID).Scan(&username, &lastSeen)
    if err != nil {
        log.Printf("Error getting user status: %v", err)
        return
    }

    status := map[string]interface{}{
        "type":      "status_update",
        "user_id":   userID,
        "username":  username,
        "is_online": isOnline,
		"last_seen":  lastSeen.Format(time.RFC3339),
        "timestamp": time.Now().Unix(),
    }

    chatMutex.RLock()
    defer chatMutex.RUnlock()

    for _, userClients := range clients {
        for _, client := range userClients {
            if err := client.Conn.WriteJSON(status); err != nil {
                // if websocket.IsUnexpectedCloseError(err) {
                //     // Queue for cleanup
                //     go func(c *Client) {
                //         unregister <- c
                //         c.Conn.Close()
                //     }(client)
                // }
            }
        }
    }
}

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
