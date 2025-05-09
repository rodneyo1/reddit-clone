package handlers

import (
	"github.com/gorilla/sessions"
)

var (
	// Session store
	Store = sessions.NewCookieStore([]byte("your-secret-key-here")) // Change this to a secure random key
)

// Initialize the store (call this from main.go)
func InitSessionStore(secretKey string) {
	Store = sessions.NewCookieStore([]byte(secretKey))
}