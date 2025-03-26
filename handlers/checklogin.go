package handlers

import (
	"encoding/json"
	"net/http"
)

func CheckLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in (e.g., by verifying a session cookie)
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		// No session cookie found, user is not logged in
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isLoggedIn": false,
		})
		return
	}

	// Verify the session ID in the database
	var userID string
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionID.Value).Scan(&userID)
	if err != nil {
		// Invalid session ID, user is not logged in
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isLoggedIn": false,
		})
		return
	}

	// User is logged in
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"isLoggedIn": true,
	})
}