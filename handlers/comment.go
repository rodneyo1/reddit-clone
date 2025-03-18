package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"
)

var PostID int

// Comment handler for processing form submissions
func CommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	postID := r.FormValue("post_id")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "Invalid post ID format", http.StatusBadRequest)
		return
	} else {
		PostID = postIDInt
	}
	content := r.FormValue("content")
	parentID := r.FormValue("parent_id") // New: Get parent comment ID if this is a reply
	userID := GetUserIdFromSession(w, r) // Fetch user ID from session

	if userID == "" {
		http.Error(w, "Please log in to comment on posts", http.StatusUnauthorized)
		return
	}

	// Validate required fields
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	if content == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if parentID != "" {
		// Convert parentID to int
		parentIDInt, err := strconv.Atoi(parentID)
		if err != nil {
			http.Error(w, "Invalid parent comment ID format", http.StatusBadRequest)
			return
		}

		// Verify that the parent comment exists
		var parentPostID int
		err = db.QueryRow("SELECT post_id FROM comments WHERE id = ?", parentIDInt).Scan(&parentPostID)
		if err == sql.ErrNoRows {
			http.Error(w, "Parent comment not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Proceed with inserting the reply since the parent comment exists
		_, err = tx.Exec(
			"INSERT INTO comments (post_id, user_id, content, parent_id, created_at) VALUES (?, ?, ?, ?, ?)",
			postIDInt, userID, content, parentIDInt, time.Now(),
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to insert comment", http.StatusInternalServerError)
			return
		}
	} else {
		// This is a top-level comment
		_, err = tx.Exec(
			"INSERT INTO comments (post_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
			postIDInt, userID, content, time.Now(),
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to insert comment", http.StatusInternalServerError)
			return
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
