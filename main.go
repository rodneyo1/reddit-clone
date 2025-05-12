package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"forum/handlers"
)

func main() {
	args := os.Args
	if len(args) != 1 {
		fmt.Println("usage: go run .")
		return
	}

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))

	// Serve HTML for home
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Serve API data for frontend fetch requests
	http.HandleFunc("/api/home", handlers.HomeHandler)
	http.HandleFunc("/api/profile", handlers.ProfileHandler)
	http.HandleFunc("/api/posts", handlers.PostHandler)
	http.HandleFunc("/api/filter", handlers.FilterHandler)
	http.HandleFunc("/api/like", handlers.LikeHandler)
	http.HandleFunc("/api/comment", handlers.CommentHandler)
	http.HandleFunc("/api/comments", handlers.GetCommentsHandler)
	http.HandleFunc("/api/comment/like", handlers.CommentLikeHandler)

	// Auth Endpoints
	http.HandleFunc("/api/register", handlers.RegisterHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)
	http.HandleFunc("/api/check-login", handlers.CheckLoginHandler)
	http.HandleFunc("/api/logout", handlers.LogoutHandler)
	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin)
	http.HandleFunc("/auth/github/login", handlers.HandleGithubLogin)
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback)
	http.HandleFunc("/auth/github/callback", handlers.HandleGithubCallback)

	// Chat API endpoints
	http.HandleFunc("/api/chats", handlers.GetChatsHandler)
	http.HandleFunc("/api/chats/create", handlers.CreateChatHandler)
	http.HandleFunc("/api/chats/messages", handlers.GetChatMessagesHandler)
	http.HandleFunc("/api/chats/members", handlers.GetChatMembersHandler)
	http.HandleFunc("/api/users/status", handlers.GetUserStatusHandler)

	// // WebSocket endpoint for real-time chat
	http.HandleFunc("/ws/chat", handlers.ChatWebSocketHandler)

	// Initialize the database and OAuth providers
	handlers.InitDB()
	handlers.InitGoogleOAuth()
	handlers.InitGithubOAuth()

	// Start the server
	log.Println("Server is running on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Helper function to identify API routes
func isAPIRoute(path string) bool {
	// Add prefixes that should not be handled by the SPA handler
	apiPrefixes := []string{
		"/api/",
		"/auth/",
		"/static/",
		"/uploads/",
		"/src/",
	}

	for _, prefix := range apiPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}
