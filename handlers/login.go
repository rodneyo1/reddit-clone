package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if credentials.Identifier == "" || credentials.Password == "" {
		respondWithError(w, "Identifier and password are required", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		respondWithError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get user with email or nickname
	var user User
	var hashedPassword string
	err = tx.QueryRow(`
        SELECT id, email, username, password 
        FROM users 
        WHERE email = ? OR nickname = ?`,
		credentials.Identifier, credentials.Identifier).Scan(
		&user.ID, &user.Email, &user.Username, &hashedPassword)

	if err == sql.ErrNoRows {
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Database error: %v", err)
		respondWithError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password)); err != nil {
		respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Delete existing sessions
	if _, err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", user.ID); err != nil {
		log.Printf("Error deleting sessions: %v", err)
		respondWithError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Create new session
	sessionID := uuid.New().String()
	expiration := time.Now().Add(24 * time.Hour)

	if _, err := tx.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, user.ID, expiration.Format(time.RFC3339),
	); err != nil {
		log.Printf("Error creating session: %v", err)
		respondWithError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		respondWithError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Set secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Return success with username
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"userID":   user.ID,
		"username": user.Username,
	})
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
