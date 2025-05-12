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

type PrivateMessage struct {
    ID         int       `json:"id"`
    SenderID   string    `json:"sender_id"`
    ReceiverID string    `json:"receiver_id"`
    Content    string    `json:"content"`
    CreatedAt  time.Time `json:"created_at"`
    IsRead     bool      `json:"is_read"`
    Sender     User      `json:"sender"`
}

type UserStatus struct {
    UserID    string    `json:"user_id"`
    IsOnline  bool      `json:"is_online"`
    LastSeen  time.Time `json:"last_seen"`
    User      User      `json:"user"`
}
