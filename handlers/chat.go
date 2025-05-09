package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for simplicity
	},
}

// Map of connected clients
var clients = make(map[string]*websocket.Conn)
var clientsMutex = &sync.Mutex{}

// GetChatsHandler returns all chats for the current user
func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	userID := GetUserIdFromSession(w,r)
	
	// Get all chat rooms for the user
	rows, err := db.Query(`
		SELECT cr.id, cr.name, cr.is_group, cr.created_at, cr.updated_at,
		       (SELECT COUNT(*) FROM chat_messages cm 
		        WHERE cm.chat_room_id = cr.id AND cm.id > crm.last_read_message_id) as unread_count
		FROM chat_rooms cr
		JOIN chat_room_members crm ON cr.id = crm.chat_room_id
		WHERE crm.user_id = ?
		ORDER BY cr.updated_at DESC
	`, userID)
	
	if err != nil {
		log.Println("Error getting chats:", err)
		http.Error(w, "Error getting chats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var chats []ChatRoom
	for rows.Next() {
		var chat ChatRoom
		var unreadCount int
		err := rows.Scan(&chat.ID, &chat.Name, &chat.IsGroup, &chat.CreatedAt, &chat.UpdatedAt, &unreadCount)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		
		chat.UnreadCount = unreadCount
		
		// Get last message
		lastMsg := db.QueryRow(`
			SELECT cm.id, cm.sender_id, u.username, cm.content, cm.sent_at
			FROM chat_messages cm
			JOIN users u ON cm.sender_id = u.id
			WHERE cm.chat_room_id = ?
			ORDER BY cm.sent_at DESC
			LIMIT 1
		`, chat.ID)
		
		var lastMessage ChatMessage
		err = lastMsg.Scan(&lastMessage.ID, &lastMessage.SenderID, &lastMessage.Username, &lastMessage.Content, &lastMessage.SentAt)
		if err == nil {
			chat.LastMessage = &lastMessage
		}
		
		chats = append(chats, chat)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"chats":   chats,
	})
}

// CreateChatHandler creates a new chat room or DM
func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get current user from session
	userID := GetUserIdFromSession(w,r)
	
	
	
	
	// Parse request
	var req struct {
		Name     string   `json:"name"`
		IsGroup  bool     `json:"is_group"`
		MemberIDs []string `json:"member_ids"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// For DMs, check if chat already exists between these two users
	if !req.IsGroup && len(req.MemberIDs) == 1 {
		var chatID int
		err = db.QueryRow(`
			SELECT cr.id
			FROM chat_rooms cr
			JOIN chat_room_members crm1 ON cr.id = crm1.chat_room_id
			JOIN chat_room_members crm2 ON cr.id = crm2.chat_room_id
			WHERE cr.is_group = 0
			AND crm1.user_id = ?
			AND crm2.user_id = ?
		`, userID, req.MemberIDs[0]).Scan(&chatID)
		
		if err == nil {
			// Chat already exists, return it
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"chat_id": chatID,
				"message": "Chat already exists",
			})
			return
		}
	}
	
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Create new chat room
	now := time.Now()
	result, err := tx.Exec(`
		INSERT INTO chat_rooms (name, is_group, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, req.Name, req.IsGroup, now, now)
	
	if err != nil {
		tx.Rollback()
		log.Println("Error creating chat room:", err)
		http.Error(w, "Error creating chat", http.StatusInternalServerError)
		return
	}
	
	chatID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Println("Error getting last insert ID:", err)
		http.Error(w, "Error creating chat", http.StatusInternalServerError)
		return
	}
	
	// Add current user as admin
	_, err = tx.Exec(`
		INSERT INTO chat_room_members (chat_room_id, user_id, is_admin)
		VALUES (?, ?, 1)
	`, chatID, userID)
	
	if err != nil {
		tx.Rollback()
		log.Println("Error adding current user to chat:", err)
		http.Error(w, "Error creating chat", http.StatusInternalServerError)
		return
	}
	
	// Add other members
	for _, memberID := range req.MemberIDs {
		_, err = tx.Exec(`
			INSERT INTO chat_room_members (chat_room_id, user_id)
			VALUES (?, ?)
		`, chatID, memberID)
		
		if err != nil {
			log.Println("Error adding member to chat:", err)
			// Continue anyway, don't fail if one member can't be added
		}
	}
	
	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		http.Error(w, "Error creating chat", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"chat_id": chatID,
	})
}

// GetChatMessagesHandler returns messages for a specific chat
func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	userID := GetUserIdFromSession(w,r)
	
	
	
	
	// Get chat ID from query params
	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "Missing chat_id parameter", http.StatusBadRequest)
		return
	}
	
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}
	
	// Check if user is a member of this chat
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM chat_room_members
		WHERE chat_room_id = ? AND user_id = ?
	`, chatID, userID).Scan(&count)
	
	if err != nil || count == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get messages
	rows, err := db.Query(`
		SELECT cm.id, cm.sender_id, u.username, u.avatar_url, cm.content, cm.sent_at
		FROM chat_messages cm
		JOIN users u ON cm.sender_id = u.id
		WHERE cm.chat_room_id = ?
		ORDER BY cm.sent_at ASC
	`, chatID)
	
	if err != nil {
		log.Println("Error getting messages:", err)
		http.Error(w, "Error getting messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Username, &msg.AvatarURL, &msg.Content, &msg.SentAt)
		if err != nil {
			log.Println("Error scanning message:", err)
			continue
		}
		msg.ChatRoomID = chatID
		messages = append(messages, msg)
	}
	
	// Update last read message
	if len(messages) > 0 {
		lastMsgID := messages[len(messages)-1].ID
		_, err = db.Exec(`
			UPDATE chat_room_members
			SET last_read_message_id = ?
			WHERE chat_room_id = ? AND user_id = ?
		`, lastMsgID, chatID, userID)
		
		if err != nil {
			log.Println("Error updating last read message:", err)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

// GetChatMembersHandler returns members of a specific chat
func GetChatMembersHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	userID := GetUserIdFromSession(w,r)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	
	// Get chat ID from query params
	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "Missing chat_id parameter", http.StatusBadRequest)
		return
	}
	
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}
	
	// Check if user is a member of this chat
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM chat_room_members
		WHERE chat_room_id = ? AND user_id = ?
	`, chatID, userID).Scan(&count)
	
	if err != nil || count == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Get members
	rows, err := db.Query(`
		SELECT crm.id, crm.user_id, u.username, u.avatar_url, crm.joined_at, crm.is_admin,
		       crm.last_read_message_id, COALESCE(us.status, 'offline') as status
		FROM chat_room_members crm
		JOIN users u ON crm.user_id = u.id
		LEFT JOIN user_statuses us ON u.id = us.user_id
		WHERE crm.chat_room_id = ?
	`, chatID)
	
	if err != nil {
		log.Println("Error getting members:", err)
		http.Error(w, "Error getting members", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var members []ChatRoomMember
	for rows.Next() {
		var member ChatRoomMember
		err := rows.Scan(
			&member.ID, &member.UserID, &member.Username, &member.AvatarURL,
			&member.JoinedAt, &member.IsAdmin, &member.LastReadMessageID, &member.Status,
		)
		if err != nil {
			log.Println("Error scanning member:", err)
			continue
		}
		member.ChatRoomID = chatID
		members = append(members, member)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"members": members,
	})
}

// GetUserStatusHandler returns the online status of users
func GetUserStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	// session, err := GetUserIdFromSession(w,r)
	// if err != nil {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	
	// Get user IDs from query params
	userIDs := r.URL.Query()["user_id"]
	if len(userIDs) == 0 {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}
	
	// Build query with placeholders
	query := `
		SELECT u.id, u.username, COALESCE(us.status, 'offline') as status, COALESCE(us.last_active, u.id) as last_active
		FROM users u
		LEFT JOIN user_statuses us ON u.id = us.user_id
		WHERE u.id IN (?`
	
	for i := 1; i < len(userIDs); i++ {
		query += ", ?"
	}
	query += ")"
	
	// Convert userIDs to interface{} for db.Query
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("Error getting user statuses:", err)
		http.Error(w, "Error getting user statuses", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var statuses []UserStatus
	for rows.Next() {
		var status UserStatus
		err := rows.Scan(&status.UserID, &status.Username, &status.Status, &status.LastActive)
		if err != nil {
			log.Println("Error scanning user status:", err)
			continue
		}
		statuses = append(statuses, status)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"statuses": statuses,
	})
}

// ChatWebSocketHandler handles WebSocket connections for real-time chat
func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	userID := GetUserIdFromSession(w,r)
	
	
	
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	
	// Register client
	clientsMutex.Lock()
	clients[userID] = conn
	clientsMutex.Unlock()
	
	// Update user status to online
	_, err = db.Exec(`
		INSERT INTO user_statuses (user_id, status, last_active)
		VALUES (?, 'online', CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
		status = 'online',
		last_active = CURRENT_TIMESTAMP
	`, userID)
	
	if err != nil {
		log.Println("Error updating user status:", err)
	}
	
	// Broadcast user status change to all clients
	broadcastStatus(userID, "online")
	
	// Handle incoming messages
	go handleWebSocketMessages(conn, userID)
}

// handleWebSocketMessages processes incoming WebSocket messages
func handleWebSocketMessages(conn *websocket.Conn, userID string) {
	defer func() {
		conn.Close()
		
		// Remove client from connected clients
		clientsMutex.Lock()
		delete(clients, userID)
		clientsMutex.Unlock()
		
		// Update user status to offline
		_, err := db.Exec(`
			UPDATE user_statuses
			SET status = 'offline', last_active = CURRENT_TIMESTAMP
			WHERE user_id = ?
		`, userID)
		
		if err != nil {
			log.Println("Error updating user status to offline:", err)
		}
		
		// Broadcast status change
		broadcastStatus(userID, "offline")
	}()
	
	for {
		// Read message from WebSocket
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			break
		}
		
		// Parse the message
		var wsMsg WebSocketMessage
		err = json.Unmarshal(msg, &wsMsg)
		if err != nil {
			log.Println("Error parsing WebSocket message:", err)
			continue
		}
		
		// Handle different message types
		switch wsMsg.Type {
		case "message":
			handleChatMessage(wsMsg, userID)
		case "status":
			handleStatusUpdate(wsMsg, userID)
		}
	}
}

// handleChatMessage processes and stores a new chat message
func handleChatMessage(wsMsg WebSocketMessage, senderID string) {
	// Parse the message data
	msgData, ok := wsMsg.Data.(map[string]interface{})
	if !ok {
		log.Println("Invalid message data format")
		return
	}
	
	content, ok := msgData["content"].(string)
	if !ok || content == "" {
		log.Println("Missing or invalid message content")
		return
	}
	
	// Store message in database
	result, err := db.Exec(`
		INSERT INTO chat_messages (chat_room_id, sender_id, content)
		VALUES (?, ?, ?)
	`, wsMsg.ChatRoomID, senderID, content)
	
	if err != nil {
		log.Println("Error storing chat message:", err)
		return
	}
	
	// Update chat room's updated_at timestamp
	_, err = db.Exec(`
		UPDATE chat_rooms
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, wsMsg.ChatRoomID)
	
	if err != nil {
		log.Println("Error updating chat room timestamp:", err)
	}
	
	// Get message ID
	msgID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting message ID:", err)
		return
	}
	
	// Get sender info
	var username, avatarURL string
	err = db.QueryRow(`
		SELECT username, avatar_url
		FROM users WHERE id = ?
	`, senderID).Scan(&username, &avatarURL)
	
	if err != nil {
		log.Println("Error getting sender info:", err)
	}
	
	// Prepare message for broadcasting
	message := ChatMessage{
		ID:         int(msgID),
		ChatRoomID: wsMsg.ChatRoomID,
		SenderID:   senderID,
		Username:   username,
		AvatarURL:  avatarURL,
		Content:    content,
		SentAt:     time.Now(),
	}
	
	// Get all members of this chat room
	rows, err := db.Query(`
		SELECT user_id
		FROM chat_room_members
		WHERE chat_room_id = ?
	`, wsMsg.ChatRoomID)
	
	if err != nil {
		log.Println("Error getting chat members:", err)
		return
	}
	defer rows.Close()
	
	// Broadcast to all members
	var members []string
	for rows.Next() {
		var memberID string
		err := rows.Scan(&memberID)
		if err != nil {
			log.Println("Error scanning member ID:", err)
			continue
		}
		members = append(members, memberID)
	}
	
	// Update sender's last read message ID
	_, err = db.Exec(`
		UPDATE chat_room_members
		SET last_read_message_id = ?
		WHERE chat_room_id = ? AND user_id = ?
	`, msgID, wsMsg.ChatRoomID, senderID)
	
	if err != nil {
		log.Println("Error updating last read message:", err)
	}
	
	// Broadcast message to all members
	broadcastToUsers(members, WebSocketMessage{
		Type:       "message",
		ChatRoomID: wsMsg.ChatRoomID,
		Data:       message,
	})
}

// handleStatusUpdate processes a status update from a user
func handleStatusUpdate(wsMsg WebSocketMessage, userID string) {
	// Parse the status data
	statusData, ok := wsMsg.Data.(map[string]interface{})
	if !ok {
		log.Println("Invalid status data format")
		return
	}
	
	status, ok := statusData["status"].(string)
	if !ok || (status != "online" && status != "offline" && status != "away") {
		log.Println("Invalid status value")
		return
	}
	
	// Update status in database
	_, err := db.Exec(`
		INSERT INTO user_statuses (user_id, status, last_active)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
		status = ?,
		last_active = CURRENT_TIMESTAMP
	`, userID, status, status)
	
	if err != nil {
		log.Println("Error updating user status:", err)
		return
	}
	
	// Broadcast status change to all clients
	broadcastStatus(userID, status)
}

// broadcastStatus sends a user's status update to all connected clients
func broadcastStatus(userID string, status string) {
	// Get user info
	var username string
	err := db.QueryRow(`
		SELECT username FROM users WHERE id = ?
	`, userID).Scan(&username)
	
	if err != nil {
		log.Println("Error getting username:", err)
		return
	}
	
	// Create status message
	statusMsg := WebSocketMessage{
		Type: "status",
		Data: UserStatus{
			UserID:     userID,
			Username:   username,
			Status:     status,
			LastActive: time.Now(),
		},
	}
	
	// Broadcast to all clients
	broadcastToAll(statusMsg)
}

// broadcastToAll sends a message to all connected clients
func broadcastToAll(msg WebSocketMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}
	
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	
	for _, conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			log.Println("Error sending message to client:", err)
			// Don't remove client here, let the read handler handle disconnects
		}
	}
}

// broadcastToUsers sends a message to specific users
func broadcastToUsers(userIDs []string, msg WebSocketMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}
	
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	
	for _, userID := range userIDs {
		if conn, ok := clients[userID]; ok {
			err := conn.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				log.Println("Error sending message to client:", err)
			}
		}
	}
}