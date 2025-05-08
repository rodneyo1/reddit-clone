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

// ChatRoom represents a chat room in the system
type ChatRoom struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatorName string    `json:"creator_name,omitempty"` // For display purposes
	Participants int      `json:"participants,omitempty"` // Count of participants
}

// ChatParticipant represents a user's participation in a chat room
type ChatParticipant struct {
	ID       int       `json:"id"`
	RoomID   int       `json:"room_id"`
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
	IsAdmin  bool      `json:"is_admin"`
	Username string    `json:"username,omitempty"` // For display purposes
}

// ChatMessage represents a message in a chat room
type ChatMessage struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username,omitempty"`  // For display purposes
	AvatarURL string    `json:"avatar_url,omitempty"` // For display purposes
	IsRead    bool      `json:"is_read,omitempty"`   // If current user has read this
}

// MessageReceipt represents a read receipt for a message
type MessageReceipt struct {
	ID        int       `json:"id"`
	MessageID int       `json:"message_id"`
	UserID    string    `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

// UnreadCount represents the count of unread messages
type UnreadCount struct {
	RoomID int `json:"room_id"`
	Count  int `json:"count"`
}