package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	// "html/template"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIdFromSession(w, r)

	// Query to fetch all posts along with user info, categories, like counts, and comments
	rows, err := db.Query(`
		SELECT p.id, p.title, p.content, p.image_path, GROUP_CONCAT(pc.category) as categories, 
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
		GROUP BY p.id, p.title, p.content, u.username, p.created_at
		ORDER BY p.created_at DESC`)
	if err != nil {
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	// Inside the HomeHandler or wherever you retrieve posts
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

		// Fetch comments for this post
		comments, err := GetCommentsForPost(post.ID)
		if err != nil {
			RenderError(w, r, "Error fetching comments", http.StatusInternalServerError)
			return
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	data := map[string]interface{}{
		"Posts":      posts,
		"IsLoggedIn": userID != "",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Error encoding JSON: %v", err)
	}
	fmt.Println(data)
	// // // Render the index page with posts
	// tmpl, err := template.ParseFiles("templates/home.html")
	// if err != nil {
	// 	RenderError(w, r, "Error parsing file", http.StatusInternalServerError)
	// 	return
	// }

	// tmpl.Execute(w, data)
}
