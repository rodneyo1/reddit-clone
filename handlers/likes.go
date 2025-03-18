package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type LikeResponse struct {
	Success      bool `json:"success"`
	LikeCount    int  `json:"like_count"`
	DislikeCount int  `json:"dislike_count"`
}

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if the user is logged in
	session, err := r.Cookie("session_id")
	if err != nil {
		// User is not logged in, return a custom JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  false,
			"error":    "You must be logged in to like a post",
			"redirect": "/login", // Add a redirect URL
		})
		return
	}

	// Get the user ID from the session
	var userID string
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", session.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Parse the form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	postID := r.FormValue("post_id")
	isLike, err := strconv.ParseBool(r.FormValue("is_like"))
	if err != nil {
		http.Error(w, "Invalid like/dislike value", http.StatusBadRequest)
		return
	}

	// Check if the user has already liked/disliked the post
	var existingIsLike bool
	err = db.QueryRow("SELECT is_like FROM likes WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&existingIsLike)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If the user is trying to toggle their like/dislike
	if err != sql.ErrNoRows {
		if existingIsLike == isLike {
			// User is trying to remove their like/dislike
			_, err = db.Exec("DELETE FROM likes WHERE post_id = ? AND user_id = ?", postID, userID)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		} else {
			// User is changing their like/dislike
			_, err = db.Exec("UPDATE likes SET is_like = ? WHERE post_id = ? AND user_id = ?", isLike, postID, userID)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}
	} else {
		// User is adding a new like/dislike
		_, err = db.Exec("INSERT INTO likes (post_id, user_id, is_like) VALUES (?, ?, ?)", postID, userID, isLike)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				http.Error(w, "You have already liked/disliked this post", http.StatusBadRequest)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	// Get the updated like and dislike counts
	var likeCount, dislikeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 1", postID).Scan(&likeCount)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 0", postID).Scan(&dislikeCount)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return the updated counts
	response := LikeResponse{
		Success:      true,
		LikeCount:    likeCount,
		DislikeCount: dislikeCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
