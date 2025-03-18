package handlers

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid" // Import UUID package
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			log.Printf("Error parsing register template: %v", err)
			RenderError(w, r, "server_error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Error executing register template: %v", err)
			RenderError(w, r, "server_error", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		email := strings.TrimSpace(r.FormValue("email"))
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Validate input
		if email == "" || username == "" || password == "" || confirmPassword == "" {
			RenderError(w, r, "invalid_input", http.StatusBadRequest)
			return
		}

		var existingUsername string
		err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUsername)
		if err == nil {
			RenderError(w, r, "Username already taken", http.StatusBadRequest)
			return
		}

		// Validate email format
		emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		if !emailRegex.MatchString(email) {
			RenderError(w, r, "invalid_email", http.StatusBadRequest)
			return
		}

		// Check password length
		if len(password) < 6 {
			RenderError(w, r, "password_too_short", http.StatusBadRequest)
			return
		}

		// Check if passwords match
		if password != confirmPassword {
			RenderError(w, r, "passwords_dont_match", http.StatusBadRequest)
			return
		}

		// Check if email already exists
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
		if err != nil {
			log.Printf("Error checking email existence: %v", err)
			RenderError(w, r, "database_error", http.StatusInternalServerError)
			return
		}
		if exists {
			RenderError(w, r, "email_exists", http.StatusConflict)
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			RenderError(w, r, "server_error", http.StatusInternalServerError)
			return
		}

		// Generate a new UUID for the user
		userID := uuid.New().String()

		// Create user
		_, err = db.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", userID, email, username, hashedPassword)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			RenderError(w, r, "database_error", http.StatusInternalServerError)
			return
		}

		// Redirect to login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	RenderError(w, r, "invalid_input", http.StatusMethodNotAllowed)
}
