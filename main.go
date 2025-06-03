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
	// Serve static files from the "static" directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))

	// Serve HTML for home
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Serve API data for frontend fetch requests
	http.HandleFunc("/api/home", handlers.HomeHandler)
	http.HandleFunc("/api/register", handlers.RegisterHandler)
	http.HandleFunc("/api/profile", handlers.ProfileHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)
	http.HandleFunc("/api/check-login", handlers.CheckLoginHandler)
	http.HandleFunc("/api/posts", handlers.PostHandler)
	http.HandleFunc("/api/logout", handlers.LogoutHandler)
	http.HandleFunc("/api/filter", handlers.FilterHandler)
	http.HandleFunc("/api/like", handlers.LikeHandler)
	http.HandleFunc("/api/comment", handlers.CommentHandler)
	http.HandleFunc("/api/comments", handlers.GetCommentsHandler)
	http.HandleFunc("/ws/chat", handlers.ChatWebsocketHandler)
	http.HandleFunc("/api/chat/users", handlers.ChatUsersHandler)
	http.HandleFunc("/api/chat/messages", handlers.ChatMessagesHandler)
	http.HandleFunc("/api/profile/update", handlers.UpdateProfileHandler)
	http.HandleFunc("/api/comment/like", handlers.CommentLikeHandler)

	// Initialize the database
	handlers.InitDB()
	go handlers.StartChatManager()

	// Start the server
	log.Println("Server is running on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}
