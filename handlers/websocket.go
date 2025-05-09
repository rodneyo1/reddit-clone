package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[int]*websocket.Conn)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	session, err := Store.Get(r, "session-name")
    if err != nil {
        log.Println("Error getting session:", err)
        return
    }

	userID, ok := session.Values["user_id"].(int)
    if !ok {
        log.Println("User not authenticated for WebSocket")
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Get user ID from session
	// session, err := Store.Get(r, "session-name")
	// if err != nil {
	// 	log.Println("Error getting session:", err)
	// 	return
	// }
	clients[userID] = conn
    log.Printf("User %d connected to WebSocket", userID)

	// userID, ok := session.Values["user_id"].(int)
	// if !ok {
	// 	log.Println("Could not get user_id from session")
	// 	return
	// }

	// Register client
	// clients[userID] = conn

	for {
		var msg PrivateMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			delete(clients, userID)
			break
		}

		// Save message to database
		_, err = db.Exec("INSERT INTO private_messages (sender_id, receiver_id, content) VALUES (?, ?, ?)",
			msg.SenderID, msg.ReceiverID, msg.Content)
		if err != nil {
			log.Println(err)
			continue
		}

		// Forward message to recipient if online
		if recipientConn, ok := clients[msg.ReceiverID]; ok {
			if err := recipientConn.WriteJSON(msg); err != nil {
				log.Println(err)
			}
		}
	}
}