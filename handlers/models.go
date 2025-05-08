package handlers

import "time"

// User represents a user in the system
type User struct {
	ID       string
	Email    string
	Username string
	Password string
}

// Post represents a post in the forum
type Post struct {
	ID             int
	UserID         string
	Title          string
	Content        string
	ImagePath      string // New field for image path
	Categories     string
	Username       string
	CreatedAt      time.Time
	CreatedAtHuman string
	LikeCount      int // Number of likes
	DislikeCount   int
	Comments       []Comment // List of comments for this post
}

type Like struct {
	ID     int
	UserID string // User who liked/disliked the post
	PostID int    // Post that was liked/disliked
	IsLike bool   // true for like, false for dislike
}

// Comment struct
type Comment struct {
	ID             int
	PostID         int
	UserID         string // Changed from int to string to match User.ID
	Content        string
	CreatedAt      time.Time // Original time
	CreatedAtHuman string    // Human-readable time
	Username       string
	ParentID       *int      // Parent comment ID, null for top-level comments
	Replies        []Comment // List of reply comments
	ReplyCount     int       // Number of replies
	LikeCount      int       // Number of likes
	DislikeCount   int       // Number of dislikes
	UserLiked      *bool     // Whether the current user liked this comment
}

// Session represents a user session
type Session struct {
	SessionID string
	UserID    string
}

// ChatRoom represents a chat room or DM conversation
type ChatRoom struct {
	ID        int       `json:"id"`
	Name      string    `json:"name,omitempty"`
	IsDM      bool      `json:"is_dm"`
	CreatedAt time.Time `json:"created_at"`
	// For client display
	DisplayName    string    `json:"display_name,omitempty"`
	LastMessage    string    `json:"last_message,omitempty"`
	LastMessageAt  time.Time `json:"last_message_at,omitempty"`
	UnreadCount    int       `json:"unread_count"`
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	ID        int       `json:"id,omitempty"`
	ChatID    int       `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Username  string    `json:"username,omitempty"`
	IsRead    bool      `json:"is_read,omitempty"`
}

// UserStatus represents a user's online status
type UserStatus struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Status     string    `json:"status"` // online, offline, away
	LastActive time.Time `json:"last_active,omitempty"`
}

// WebSocketMessage represents a message sent through websocket
type WebSocketMessage struct {
	Type    string      `json:"type"` // message, status_update, typing
	Payload interface{} `json:"payload"`
}