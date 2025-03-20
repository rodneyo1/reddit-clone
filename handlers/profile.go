package handlers

import (
	// "html/template"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIdFromSession(w, r)
	if userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's created posts
	createdPosts, err := db.Query(`
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			p.image_path,
			GROUP_CONCAT(DISTINCT pc.category) as categories, 
			u.username, 
			p.created_at,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislike_count
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		LEFT JOIN post_categories pc ON p.id = pc.post_id 
		WHERE p.user_id = ?
		GROUP BY p.id 
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		log.Printf("Error fetching user's posts: %v", err)
		RenderError(w, r, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer createdPosts.Close()

	var userPosts []Post
	for createdPosts.Next() {
		var post Post
		var createdAt time.Time
		var categories string
		err := createdPosts.Scan(
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
			continue
		}

		// Set the CreatedAt field and the human-readable time
		post.CreatedAt = createdAt
		post.CreatedAtHuman = TimeAgo(createdAt)

		post.Categories = categories
		userPosts = append(userPosts, post)
	}

	// Get user's liked posts
	likedPosts, err := db.Query(`
		SELECT 
			p.id, 
			p.title, 
			p.content,
			p.image_path, 
			GROUP_CONCAT(DISTINCT pc.category) as categories, 
			u.username, 
			p.created_at,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1) as like_count,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0) as dislike_count
		FROM posts p 
		JOIN users u ON p.user_id = u.id 
		LEFT JOIN post_categories pc ON p.id = pc.post_id 
		JOIN likes l ON p.id = l.post_id
		WHERE l.user_id = ? AND l.is_like = 1
		GROUP BY p.id 
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		log.Printf("Error fetching liked posts: %v", err)
		RenderError(w, r, "Error fetching liked posts", http.StatusInternalServerError)
		return
	}
	defer likedPosts.Close()

	var userLikedPosts []Post
	for likedPosts.Next() {
		var post Post
		var createdAt time.Time
		var categories string
		err := likedPosts.Scan(
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
			log.Printf("Error scanning liked post: %v", err)
			continue
		}

		// Set the CreatedAt field and the human-readable time
		post.CreatedAt = createdAt
		post.CreatedAtHuman = TimeAgo(createdAt)

		post.Categories = categories
		userLikedPosts = append(userLikedPosts, post)
	}

	// Get user information
	var username string
	var email string
	err = db.QueryRow("SELECT username, email FROM users WHERE id = ?", userID).Scan(&username, &email)
	if err != nil {
		log.Printf("Error fetching user info: %v", err)
		RenderError(w, r, "Error fetching user information", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Username":     username,
		"Email":        email,
		"CreatedPosts": userPosts,
		"LikedPosts":   userLikedPosts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
	// tmpl, err := template.ParseFiles("templates/profile.html")
	// if err != nil {
	// 	log.Printf("Error parsing profile template: %v", err)
	// 	RenderError(w, r, "Error loading profile page", http.StatusInternalServerError)
	// 	return
	// }

	// if err := tmpl.Execute(w, data); err != nil {
	// 	log.Printf("Error executing profile template: %v", err)
	// 	RenderError(w, r, "Error rendering profile page", http.StatusInternalServerError)
	// 	return
	// }
}
