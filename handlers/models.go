package handlers

import "time"

// User represents a user in the system
type User struct {
    ID        string
    Email     string
    Username  string
    Password  string
    GoogleID  string    `json:"google_id,omitempty"`
    GithubID  string    `json:"github_id,omitempty"`
    AvatarURL string    `json:"avatar_url,omitempty"`
    Nickname  string    `json:"nickname,omitempty"`
    Age       int       `json:"age,omitempty"`
    Gender    string    `json:"gender,omitempty"`
    FirstName string    `json:"first_name,omitempty"`
    LastName  string    `json:"last_name,omitempty"`
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

// ChatRoom represents a chat conversation between users
type ChatRoom struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	IsGroup   bool      `json:"is_group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastMessage *ChatMessage `json:"last_message,omitempty"`
	UnreadCount int     `json:"unread_count,omitempty"`
	Members    []ChatRoomMember `json:"members,omitempty"`
}

// ChatRoomMember represents a user in a chat room
type ChatRoomMember struct {
	ID             int       `json:"id"`
	ChatRoomID     int       `json:"chat_room_id"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	AvatarURL      string    `json:"avatar_url"`
	JoinedAt       time.Time `json:"joined_at"`
	IsAdmin        bool      `json:"is_admin"`
	LastReadMessageID int    `json:"last_read_message_id"`
	Status         string    `json:"status,omitempty"`
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	ID         int       `json:"id"`
	ChatRoomID int       `json:"chat_room_id"`
	SenderID   string    `json:"sender_id"`
	Username   string    `json:"username,omitempty"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sent_at"`
}

// UserStatus represents the online status of a user
type UserStatus struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username,omitempty"`
	Status     string    `json:"status"` // online, offline, away
	LastActive time.Time `json:"last_active"`
}

// WebSocketMessage represents a message sent via WebSocket
type WebSocketMessage struct {
	Type       string      `json:"type"`  // message, status, notification, etc.
	ChatRoomID int         `json:"chat_room_id,omitempty"`
	Data       interface{} `json:"data"`
}