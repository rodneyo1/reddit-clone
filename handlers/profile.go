package handlers

import (
	// "html/template"
	"encoding/json"
	"fmt"
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
	var user User
	err = db.QueryRow(`
    SELECT username, email,
           COALESCE(nickname, ''),
           COALESCE(avatar_url, ''),
           COALESCE(age, 0),
           COALESCE(gender, ''),
           COALESCE(first_name, ''),
           COALESCE(last_name, '')
    FROM users WHERE id = ?`, userID).
		Scan(
			&user.Username,
			&user.Email,
			&user.Nickname,
			&user.AvatarURL,
			&user.Age,
			&user.Gender,
			&user.FirstName,
			&user.LastName,
		)

	if err != nil {
		log.Printf("Error fetching user info: %v", err)
		RenderError(w, r, "Error fetching user information", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Username":     user.Username,
		"Email":        user.Email,
		"Nickname":     user.Nickname,
		"AvatarURL":    user.AvatarURL,
		"Age":          user.Age,
		"Gender":       user.Gender,
		"FirstName":    user.FirstName,
		"LastName":     user.LastName,
		"CreatedPosts": userPosts,
		"LikedPosts":   userLikedPosts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
	fmt.Println(data)
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

// UpdateProfileHandler handles the profile update requests
// It supports both POST and PUT methods for updating user information
// It expects a JSON payload with the following fields:
// - id: User ID (required)
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	userID := GetUserIdFromSession(w, r)
if userID == "" {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return
}


	var payload struct {
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatar_url"`
		Age       int   `json:"age"` // pointer to distinguish 0 from missing
		Gender    string `json:"gender"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	fmt.Println(payload)

	query := `UPDATE users SET 
		nickname = COALESCE(NULLIF(?, ''), nickname),
		avatar_url = COALESCE(NULLIF(?, ''), avatar_url),
		age = COALESCE(?, age),
		gender = COALESCE(NULLIF(?, ''), gender),
		first_name = COALESCE(NULLIF(?, ''), first_name),
		last_name = COALESCE(NULLIF(?, ''), last_name)
		WHERE id = ?`

	_, err := db.Exec(query,
		payload.Nickname,
		payload.AvatarURL,
		payload.Age,
		payload.Gender,
		payload.FirstName,
		payload.LastName,
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Profile updated successfully",
	})
}
