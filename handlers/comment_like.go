package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// CommentLikeResponse is the response for like/dislike actions
type CommentLikeResponse struct {
	LikeCount    int   `json:"likeCount"`
	DislikeCount int   `json:"dislikeCount"`
	UserLiked    *bool `json:"userLiked"`
}

// CommentLikeHandler handles liking/disliking comments
func CommentLikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from session
	userID := GetUserIdFromSession(w, r)
	if userID == "" {
		http.Error(w, "Please log in to like or dislike comments", http.StatusUnauthorized)
		return
	}

	// Parse comment ID and like status from request
	commentID := r.FormValue("comment_id")
	isLike := r.FormValue("is_like")

	if commentID == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	commentIDInt, err := strconv.Atoi(commentID)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Verify comment exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)", commentIDInt).Scan(&exists)
	if err != nil {
		log.Printf("Error checking comment existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	isLikeBool := isLike == "true"

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check if user has already liked/disliked this comment
	var existingIsLike sql.NullBool
	err = tx.QueryRow(
		"SELECT is_like FROM comment_likes WHERE comment_id = ? AND user_id = ?",
		commentIDInt, userID,
	).Scan(&existingIsLike)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking existing like: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingIsLike.Valid {
		if existingIsLike.Bool == isLikeBool {
			// Remove the like/dislike if clicking the same button
			_, err = tx.Exec("DELETE FROM comment_likes WHERE comment_id = ? AND user_id = ?",
				commentIDInt, userID)
		} else {
			// Update from like to dislike or vice versa
			_, err = tx.Exec("UPDATE comment_likes SET is_like = ? WHERE comment_id = ? AND user_id = ?",
				isLikeBool, commentIDInt, userID)
		}
	} else {
		// Add new like/dislike
		_, err = tx.Exec("INSERT INTO comment_likes (comment_id, user_id, is_like) VALUES (?, ?, ?)",
			commentIDInt, userID, isLikeBool)
	}

	if err != nil {
		log.Printf("Error updating like status: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get updated counts and user's current like status
	var response CommentLikeResponse
	var userLiked sql.NullBool
	err = tx.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_like = 1),
			(SELECT COUNT(*) FROM comment_likes WHERE comment_id = ? AND is_like = 0),
			CASE 
				WHEN EXISTS (SELECT 1 FROM comment_likes WHERE comment_id = ? AND user_id = ?)
				THEN (SELECT is_like FROM comment_likes WHERE comment_id = ? AND user_id = ?)
				ELSE NULL 
			END
	`, commentIDInt, commentIDInt, commentIDInt, userID, commentIDInt, userID).Scan(&response.LikeCount, &response.DislikeCount, &userLiked)

	if err != nil {
		log.Printf("Error getting updated counts: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if userLiked.Valid {
		response.UserLiked = &userLiked.Bool
	} else {
		response.UserLiked = nil
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
