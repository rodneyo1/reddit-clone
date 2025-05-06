package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

var PostID int

// CommentHandler handles POST requests for creating new comments
func CommentHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var request struct {
		PostID   int    `json:"post_id"`
		Content  string `json:"content"`
		ParentID *int   `json:"parent_id,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("JSON parse error:", err)
		http.Error(w, `{"error":"Invalid request format"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.PostID <= 0 {
		http.Error(w, `{"error":"Invalid post ID"}`, http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, `{"error":"Comment content cannot be empty"}`, http.StatusBadRequest)
		return
	}

	// Get user from session
	userID := GetUserIdFromSession(w, r)
	if userID == "" {
		http.Error(w, `{"error":"Please log in to comment"}`, http.StatusUnauthorized)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Database begin error:", err)
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if request.ParentID != nil {
		// Verify that the parent comment exists and belongs to the same post
		var parentPostID int
		err = tx.QueryRow("SELECT post_id FROM comments WHERE id = ?", *request.ParentID).Scan(&parentPostID)
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Parent comment not found"}`, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("Parent comment check error:", err)
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}

		if parentPostID != request.PostID {
			http.Error(w, `{"error":"Parent comment doesn't belong to this post"}`, http.StatusBadRequest)
			return
		}

		// Insert reply
		_, err = tx.Exec(
			"INSERT INTO comments (post_id, user_id, content, parent_id, created_at) VALUES (?, ?, ?, ?, ?)",
			request.PostID, userID, request.Content, *request.ParentID, time.Now(),
		)
	} else {
		// Insert top-level comment
		_, err = tx.Exec(
			"INSERT INTO comments (post_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
			request.PostID, userID, request.Content, time.Now(),
		)
	}

	if err != nil {
		log.Println("Insert comment error:", err)
		http.Error(w, `{"error":"Failed to save comment"}`, http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Println("Commit error:", err)
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Comment posted successfully",
	})
}

// GetCommentsHandler handles GET requests for fetching comments
func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
    w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
        return
    }


	// Get post_id from query params
	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Invalid post ID"})
        return
    }

	comments, err := GetCommentsForPost(postID)
    if err != nil {
        log.Printf("Error fetching comments for post %d: %v", postID, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load comments"})
        return
    }


	 // Return comments
	 if err := json.NewEncoder(w).Encode(comments); err != nil {
        log.Printf("Error encoding comments: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to format response"})
    }
}

func jsonError(w http.ResponseWriter, message string, statusCode int) {
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{
        "error": message,
    })
}
// Fetch comments for a specific post
var GetCommentsForPost = func(postID int) ([]Comment, error) {
	// First, get all comments for this post
	rows, err := db.Query(`
		SELECT 
			c.id, 
			c.post_id,
			c.user_id,
			c.content,
			c.created_at,
			u.username,
			c.parent_id,
			(SELECT COUNT(*) FROM comments r WHERE r.parent_id = c.id) as reply_count,
			(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id AND cl.is_like = 1) as like_count,
			(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id AND cl.is_like = 0) as dislike_count
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ? AND c.parent_id IS NULL
		ORDER BY c.created_at DESC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var createdAt time.Time
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&createdAt,
			&comment.Username,
			&comment.ParentID,
			&comment.ReplyCount,
			&comment.LikeCount,
			&comment.DislikeCount,
		)
		if err != nil {
			return nil, err
		}

		// Set the CreatedAt field and the human-readable time
		comment.CreatedAt = createdAt
		comment.CreatedAtHuman = TimeAgo(createdAt)

		// Get replies for this comment
		replies, err := GetCommentReplies(comment.ID)
		if err != nil {
			return nil, err
		}
		comment.Replies = replies

		comments = append(comments, comment)
	}

	return comments, nil
}

// Get replies for a specific comment
var GetCommentReplies = func(commentID int) ([]Comment, error) {
	rows, err := db.Query(`
		SELECT 
			c.id, 
			c.post_id,
			c.user_id,
			c.content,
			c.created_at,
			u.username,
			c.parent_id,
			0 as reply_count,
			(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id AND cl.is_like = 1) as like_count,
			(SELECT COUNT(*) FROM comment_likes cl WHERE cl.comment_id = c.id AND cl.is_like = 0) as dislike_count
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.parent_id = ?
		ORDER BY c.created_at ASC
	`, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []Comment
	for rows.Next() {
		var reply Comment
		var createdAt time.Time
		err := rows.Scan(
			&reply.ID,
			&reply.PostID,
			&reply.UserID,
			&reply.Content,
			&createdAt,
			&reply.Username,
			&reply.ParentID,
			&reply.ReplyCount,
			&reply.LikeCount,
			&reply.DislikeCount,
		)
		if err != nil {
			return nil, err
		}

		// Set the CreatedAt field and the human-readable time
		reply.CreatedAt = createdAt
		reply.CreatedAtHuman = TimeAgo(createdAt)

		replies = append(replies, reply)
	}

	return replies, nil
}

// Get user ID from session
var GetUserIdFromSession = func(w http.ResponseWriter, r *http.Request) string {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}

	var userID string
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
	if err == sql.ErrNoRows {
		// Clear invalid session
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
		})
		return ""
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return ""
	}

	return userID
}

// Fetch a single post by ID
func GetPostByID(id string) (Post, error) {
	var post Post
	err := db.QueryRow("SELECT id, title, content FROM posts WHERE id = ?", id).Scan(&post.ID, &post.Title, &post.Content)
	return post, err
}
