package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var validCategories = []string{
	"technology",
	"general",
	"lifestyle",
	"entertainment",
	"gaming",
	"food",
	"business",
	"religion",
	"health",
	"music",
	"sports",
	"beauty",
	"jobs",
}

func isValidCategory(category string) bool {
	for _, validCategory := range validCategories {
		if validCategory == category {
			return true
		}
	}
	return false
}

func FilterHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in
	var userID string
	sessionCookie, err := r.Cookie("session_id")
	isLoggedIn := false // Flag to check if the user is logged in

	if err == nil {
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
		if err == nil {
			isLoggedIn = true // User is logged in
		} else if err == sql.ErrNoRows {
			// Clear the invalid session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})
		} else {
			log.Printf("Database error: %v", err)
			RenderError(w, r, "Database Error", http.StatusInternalServerError)
			return
		}
	}

	// Get the category from the query parameters
	category := r.URL.Query().Get("category")
	// if category == "" {
	// 	RenderError(w, r, "Error: Category parameter is missing or incomplete.", http.StatusBadRequest)
	// 	return
	// }

	// Validate the category
	if category != "all" && category != "" && !isValidCategory(category) {
		RenderError(w, r, "Invalid category selected", http.StatusBadRequest)
		return
	}

	// Query to fetch posts based on the selected category
	query := `
		SELECT p.id, p.title, p.content, p.image_path, GROUP_CONCAT(DISTINCT pc.category) as categories, 
		u.username, p.created_at, 
		COALESCE(l.like_count, 0) AS like_count,
		COALESCE(l.dislike_count, 0) AS dislike_count
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN (
			SELECT post_id, 
			COUNT(CASE WHEN is_like = 1 THEN 1 END) AS like_count,
			COUNT(CASE WHEN is_like = 0 THEN 1 END) AS dislike_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
	`
	if category != "all" && category != "" {
		query += " WHERE pc.category = ?"
	}
	query += " GROUP BY p.id, p.title, p.content, u.username, p.created_at ORDER BY p.created_at DESC"

	// Execute the query
	var rows *sql.Rows
	if category != "all" && category != "" {
		rows, err = db.Query(query, category)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse the rows into a slice of Post structs
	var posts []Post
	for rows.Next() {
		var post Post
		var createdAt time.Time
		var categories sql.NullString
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&categories,
			&post.Username,
			&createdAt,
			&post.LikeCount,
			&post.DislikeCount,
		)
		if err != nil {
			log.Printf("Error scanning post: %v", err)
			RenderError(w, r, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		if categories.Valid {
			post.Categories = categories.String
		} else {
			post.Categories = ""
		}

		// Set the CreatedAt field and the human-readable time
		post.CreatedAt = createdAt
		post.CreatedAtHuman = TimeAgo(createdAt)

		commentQuery := `
			SELECT c.id, c.content, u.username, c.created_at, 
       COALESCE(clike.like_count, 0) AS like_count,
       COALESCE(cdislike.dislike_count, 0) AS dislike_count
FROM comments c
JOIN users u ON c.user_id = u.id
LEFT JOIN (
    SELECT comment_id, COUNT(CASE WHEN is_like = 1 THEN 1 END) AS like_count
    FROM comment_likes
    GROUP BY comment_id
) clike ON c.id = clike.comment_id
LEFT JOIN (
    SELECT comment_id, COUNT(CASE WHEN is_like = 0 THEN 1 END) AS dislike_count
    FROM comment_likes
    GROUP BY comment_id
) cdislike ON c.id = cdislike.comment_id
WHERE c.post_id = ?
ORDER BY c.created_at DESC

		`
		commentRows, err := db.Query(commentQuery, post.ID)
		if err != nil {
			log.Printf("Error fetching comments: %v", err)
			RenderError(w, r, "Error fetching comments", http.StatusInternalServerError)
			return
		}
		defer commentRows.Close()

		var comments []Comment
		for commentRows.Next() {
			var comment Comment
			var createdAt time.Time
			err := commentRows.Scan(&comment.ID, &comment.Content, &comment.Username, &createdAt, &comment.LikeCount, &comment.DislikeCount)
			if err != nil {
				log.Printf("Error scanning comment: %v", err)
				RenderError(w, r, "Error scanning comments", http.StatusInternalServerError)
				return
			}
			comments = append(comments, comment)
		}

		// Set the CreatedAt field and the human-readable time
		post.CreatedAt = createdAt
		post.CreatedAtHuman = TimeAgo(createdAt)

		post.Comments = comments
		posts = append(posts, post)
	}

	data := map[string]interface{}{
		"Posts":            posts,
		"IsLoggedIn":       isLoggedIn,
		"SelectedCategory": category,
	}
	// fmt.Println(data)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Error encoding JSON: %v", err)
	}
}
