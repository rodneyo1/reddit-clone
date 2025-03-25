package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	// API endpoints that return JSON data
	http.HandleFunc("/api/home", handlers.HomeHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)
	http.HandleFunc("/api/filter", handlers.FilterHandler)
	http.HandleFunc("/api/profile", handlers.ProfileHandler)
	
	// API endpoints for actions (post, like, comment)
	http.HandleFunc("/api/post", handlers.PostHandler)
	http.HandleFunc("/api/like", handlers.LikeHandler)
	http.HandleFunc("/api/comment", handlers.CommentHandler)
	http.HandleFunc("/api/comment/like", handlers.CommentLikeHandler)

	// Authentication endpoints
	http.HandleFunc("/auth/login", handlers.LoginHandler)
	http.HandleFunc("/auth/register", handlers.RegisterHandler)
	http.HandleFunc("/auth/logout", handlers.LogoutHandler)
	
	// OAuth routes
	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback)
	http.HandleFunc("/auth/github/login", handlers.HandleGithubLogin)
	http.HandleFunc("/auth/github/callback", handlers.HandleGithubCallback)

	// SPA handler for all other routes - serves index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Exclude API and auth routes
		if isAPIRoute(r.URL.Path) {
			http.NotFound(w, r)
			return
		}
		
		// Serve the SPA index.html for all frontend routes
		loginPath := filepath.Join("static", "index.html")
		http.ServeFile(w, r, loginPath)
	})

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