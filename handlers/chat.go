// chat.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn     *websocket.Conn
	username string
	userID   string
}

var clients = make(map[*Client]bool)

type WsMessage struct {
	Type      string    `json:"type"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func authenticateUser(r *http.Request) (string, string, bool) {
	// Get session cookie from request
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("No session cookie found")
		return "", "", false
	}

	// Look up session in database
	var userID, username string
	err = db.QueryRow(`
		SELECT u.id, u.username 
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.session_id = ?
		AND s.expires_at > ?
	`, cookie.Value, time.Now().Format("2006-01-02 15:04:05")).Scan(&userID, &username)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Invalid or expired session")
		} else {
			log.Printf("Database error: %v", err)
		}
		return "", "", false
	}

	return userID, username, true
}

func ChatWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	userID, username, ok := authenticateUser(r)
	if !ok {
		log.Println("Authentication failed for WebSocket connection")
		return
	}

	client := &Client{
		conn:     conn,
		username: username,
		userID:   userID,
	}
	clients[client] = true
	defer delete(clients, client)

	updateUserStatus(userID, true)
	defer updateUserStatus(userID, false)

	notifyPresence()

	for {
		var msg WsMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		msg.Timestamp = time.Now()
		msg.Sender = userID

		if err := saveMessageToDB(msg); err != nil {
			log.Printf("Error saving message: %v", err)
		}

		broadcastMessage(msg)
	}
}

func ChatUsersHandler(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := authenticateUser(r); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := getUsersFromDB()
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := make([]struct {
		User
		Online bool `json:"online"`
	}, len(users))

	for i, user := range users {
		response[i] = struct {
			User
			Online bool `json:"online"`
		}{
			User:   user,
			Online: isUserOnline(user.ID),
		}
	}

	json.NewEncoder(w).Encode(response)
}

func ChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := authenticateUser(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	recipient := r.URL.Query().Get("recipient")
	offset := r.URL.Query().Get("offset")

	messages, err := getMessagesFromDB(userID, recipient, offset)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(messages)
}

// Database operations
func getUsersFromDB() ([]User, error) {
	rows, err := db.Query("SELECT id, username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username); err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func getMessagesFromDB(sender, recipient, offset string) ([]PrivateMessage, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, created_at, is_read 
		FROM messages 
		WHERE (sender_id = ? AND receiver_id = ?)
		OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at DESC
		LIMIT 10 OFFSET ?`
	
	rows, err := db.Query(query, sender, recipient, recipient, sender, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []PrivateMessage
	for rows.Next() {
		var pm PrivateMessage
		if err := rows.Scan(&pm.ID, &pm.SenderID, &pm.ReceiverID, &pm.Content, &pm.CreatedAt, &pm.IsRead); err != nil {
			continue
		}
		messages = append(messages, pm)
	}
	return messages, nil
}

func saveMessageToDB(msg WsMessage) error {
	_, err := db.Exec(
		"INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)",
		msg.Sender,
		msg.Recipient,
		msg.Content,
	)
	return err
}

func isUserOnline(userID string) bool {
	for client := range clients {
		if client.userID == userID {
			return true
		}
	}
	return false
}

func updateUserStatus(userID string, online bool) {
	_, err := db.Exec(
		"UPDATE user_status SET is_online = ?, last_seen = ? WHERE user_id = ?",
		online,
		time.Now(),
		userID,
	)
	if err != nil {
		log.Printf("Error updating user status: %v", err)
	}
}

func notifyPresence() {
	users := make([]struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Online   bool   `json:"online"`
	}, 0, len(clients))

	for client := range clients {
		users = append(users, struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Online   bool   `json:"online"`
		}{
			ID:       client.userID,
			Username: client.username,
			Online:   true,
		})
	}

	msg := struct {
		Type  string      `json:"type"`
		Users interface{} `json:"users"`
	}{
		Type:  "presence",
		Users: users,
	}

	for client := range clients {
		if err := client.conn.WriteJSON(msg); err != nil {
			client.conn.Close()
			delete(clients, client)
		}
	}
}

func broadcastMessage(msg WsMessage) {
	for client := range clients {
		if client.userID == msg.Recipient {
			if err := client.conn.WriteJSON(msg); err != nil {
				client.conn.Close()
				delete(clients, client)
			}
		}
	}
}