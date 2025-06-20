package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request method",
		})
		return
	}

	// Parse request body
	var newUser struct {
		Email     string `json:"email"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatar_url"`
		Age       int    `json:"age"`
		Gender    string `json:"gender"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	// DECODE THE JSON REQUEST BODY FIRST
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON format",
		})
		return
	}

	// NOW check if required fields are empty
	if newUser.Email == "" ||
		newUser.Username == "" ||
		newUser.Password == "" ||
		newUser.Nickname == "" ||
		newUser.Age == 0 ||
		newUser.Gender == "" ||
		newUser.FirstName == "" ||
		newUser.LastName == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "All fields are required",
		})
		return
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(newUser.Email) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid email format",
		})
		return
	}

	// Check password length
	if len(newUser.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Password must be at least 6 characters long",
		})
		return
	}

	if newUser.Age < 1 || newUser.Age > 120 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid Age",
		})
		return
	}

	// Check if username already exists
	var existingUsername string
	err := db.QueryRow("SELECT username FROM users WHERE username = ?", newUser.Username).Scan(&existingUsername)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Username already taken",
		})
		return
	}

	// Check if email already exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", newUser.Email).Scan(&exists)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Database error",
		})
		return
	}
	if exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Server error",
		})
		return
	}

	// Generate a new UUID for the user
	userID := uuid.New().String()
	// Generate default avatar using robohash
	newUser.AvatarURL = "https://robohash.org/" + userID

	// Create user
	_, err = db.Exec(
		"INSERT INTO users (id, email, username, password, nickname, first_name, last_name, age, gender, avatar_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		userID, newUser.Email, newUser.Username, hashedPassword, newUser.Nickname, newUser.FirstName, newUser.LastName, newUser.Age, newUser.Gender, newUser.AvatarURL,
	)
	if err != nil {
		fmt.Println("Database insert error:", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}