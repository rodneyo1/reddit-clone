package handlers

import (
	"database/sql"
	// "fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxImageSize = 20 * 1024 * 1024 // 20 MB

func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Check if the user is logged in
	var userID string
	sessionCookie, err := r.Cookie("session_id")
	if err == nil {
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
		if err == sql.ErrNoRows {
			// Clear the invalid session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})
		} else if err != nil {
			log.Printf("Database error: %v", err)
			RenderError(w, r, "Database Error", http.StatusInternalServerError)
			return
		}
	}

	// Handle POST request (create a new post)
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categories := r.Form["category"] // Get multiple categories

	// Validate input
	if title == "" || content == "" || len(categories) == 0 {
		RenderError(w, r, "Title, content, and at least one category are required", http.StatusBadRequest)
		return
	}

	// Handle image upload
	var imagePath string
	if r.MultipartForm != nil {
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Check file size
			if header.Size > maxImageSize {
				RenderError(w, r, "Image size exceeds 20 MB limit", http.StatusBadRequest)
				return
			}

			// Validate image type
			ext := strings.ToLower(filepath.Ext(header.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
				RenderError(w, r, "Invalid image type. Only JPEG, PNG, and GIF are allowed.", http.StatusBadRequest)
				return
			}

			// Ensure the uploads directory exists
			if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
				RenderError(w, r, "Error creating uploads directory", http.StatusInternalServerError)
				return
			}

			// Save the image to the uploads directory
			imagePath = "uploads/" + header.Filename
			out, err := os.Create(imagePath)
			if err != nil {
				log.Printf("Error saving image: %v", err)
				RenderError(w, r, "Error saving image", http.StatusInternalServerError)
				return
			}
			defer out.Close()
			if _, err := io.Copy(out, file); err != nil {
				log.Printf("Error writing image to file: %v", err)
				RenderError(w, r, "Error writing image to file", http.StatusInternalServerError)
				return
			}
		}
	}

	// Insert the new post into the database
	result, err := db.Exec("INSERT INTO posts (user_id, title, content, image_path, created_at) VALUES (?, ?, ?, ?, ?)", userID, title, content, imagePath, time.Now())
	if err != nil {
		log.Printf("Error creating post: %v", err)
		RenderError(w, r, "Error creating post", http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly created post
	postID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error retrieving post ID: %v", err)
		RenderError(w, r, "Error retrieving post ID", http.StatusInternalServerError)
		return
	}

	// Insert categories into the database
	for _, category := range categories {
		_, err = db.Exec("INSERT INTO post_categories (post_id, category) VALUES (?, ?)", postID, category)
		if err != nil {
			log.Printf("Error inserting category: %v", err)
			RenderError(w, r, "Error inserting categories", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Post created successfully"}`))
}
