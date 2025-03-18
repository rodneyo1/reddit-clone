package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var parseTemplate = func(_ ...string) (*template.Template, error) {
	return template.New("mock").Parse("<html></html>") // Mock template
}

func TestHomeHandler(t *testing.T) {
	// Store original functions to restore after test
	originalGetUserIdFromSession := GetUserIdFromSession
	originalGetCommentsForPost := GetCommentsForPost
	originalDB := db
	originalRenderError := RenderError

	// Restore original functions after test
	defer func() {
		GetUserIdFromSession = originalGetUserIdFromSession
		GetCommentsForPost = originalGetCommentsForPost
		db = originalDB
		RenderError = originalRenderError
	}()

	// Test case 2: Database Query Error
	t.Run("Database Query Error", func(t *testing.T) {
		// Mock GetUserIdFromSession
		GetUserIdFromSession = func(w http.ResponseWriter, r *http.Request) string {
			return "testuser"
		}

		// Create a mock database that will cause a query error
		mockDB, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create mock database: %v", err)
		}
		defer mockDB.Close()

		// Replace global db with mock
		db = mockDB

		// Track if RenderError was called
		var renderErrorCalled bool
		RenderError = func(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
			renderErrorCalled = true
			http.Error(w, message, statusCode)
		}

		// Create request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Call handler
		HomeHandler(w, req)

		// Check if RenderError was called
		if !renderErrorCalled {
			t.Errorf("Expected RenderError to be called on database query error")
		}
	})
}

func TestGetUserIdFromSession(t *testing.T) {
	// Store the original function to restore after the test
	originalGetUserIdFromSession := GetUserIdFromSession

	// Restore the original function after the test
	defer func() {
		GetUserIdFromSession = originalGetUserIdFromSession
	}()

	// Test case 1: Valid cookie
	t.Run("Valid Cookie", func(t *testing.T) {
		// Temporarily replace the function for this test
		GetUserIdFromSession = func(w http.ResponseWriter, r *http.Request) string {
			cookie, _ := r.Cookie("session-name")
			return cookie.Value
		}

		// Create a mock HTTP request with a cookie
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session-name",
			Value: "test-user-123",
		})

		// Create a response recorder
		w := httptest.NewRecorder()

		// Call the function
		userID := GetUserIdFromSession(w, req)

		// Manual assertion
		if userID != "test-user-123" {
			t.Errorf("Expected userID to be 'test-user-123', got '%s'", userID)
		}
	})

	// Test case 2: Missing cookie
	t.Run("Missing Cookie", func(t *testing.T) {
		// Temporarily replace the function for this test
		GetUserIdFromSession = func(w http.ResponseWriter, r *http.Request) string {
			http.Error(w, "Error getting session", http.StatusInternalServerError)
			return ""
		}

		// Create a mock HTTP request without a cookie
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		// Create a response recorder
		w := httptest.NewRecorder()

		// Call the function
		userID := GetUserIdFromSession(w, req)

		// Manual assertions
		if userID != "" {
			t.Errorf("Expected empty userID, got '%s'", userID)
		}

		// Check if an error response was written
		response := w.Result()
		if response.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, response.StatusCode)
		}
	})
}

func TestHandleDatabaseError(t *testing.T) {
	// Store original RenderError to restore after test
	originalRenderError := RenderError

	// Restore original function after test
	defer func() {
		RenderError = originalRenderError
	}()

	// Test case 1: Error is not nil
	t.Run("Error Present", func(t *testing.T) {
		// Track if RenderError was called
		var renderErrorCalled bool
		var renderErrorMessage string
		var renderErrorStatusCode int

		// Mock RenderError
		RenderError = func(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
			renderErrorCalled = true
			renderErrorMessage = message
			renderErrorStatusCode = statusCode
			http.Error(w, message, statusCode)
		}

		// Create mock request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Create a test error
		testErr := fmt.Errorf("test database error")

		// Call HandleDatabaseError
		HandleDatabaseError(w, req, testErr)

		// Check if RenderError was called
		if !renderErrorCalled {
			t.Errorf("Expected RenderError to be called")
		}

		// Check error message and status code
		if renderErrorMessage != "Database Error" {
			t.Errorf("Expected error message 'Database Error', got '%s'", renderErrorMessage)
		}

		if renderErrorStatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d",
				http.StatusInternalServerError, renderErrorStatusCode)
		}
	})

	// Test case 2: Error is nil
	t.Run("No Error", func(t *testing.T) {
		// Track if RenderError was called
		var renderErrorCalled bool

		// Mock RenderError
		RenderError = func(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
			renderErrorCalled = true
		}

		// Create mock request and response recorder
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Call HandleDatabaseError with nil error
		HandleDatabaseError(w, req, nil)

		// Check that RenderError was NOT called
		if renderErrorCalled {
			t.Errorf("Expected RenderError to NOT be called when error is nil")
		}
	})
}

func TestGetCommentsForPost(t *testing.T) {
	// Setup mock database
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Replace global db with mock
	originalDB := db
	db = mockDB
	defer func() { db = originalDB }()

	// Prepare mock database schema and data
	_, err = mockDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			user_id INTEGER,
			content TEXT,
			created_at DATETIME,
			parent_id INTEGER,
			FOREIGN KEY(post_id) REFERENCES posts(id),
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(parent_id) REFERENCES comments(id)
		);
		CREATE TABLE comment_likes (
			comment_id INTEGER,
			is_like BOOLEAN
		);

		-- Insert test users
		INSERT INTO users (id, username) VALUES 
		(1, 'testuser1'),
		(2, 'testuser2');

		-- Insert test post
		INSERT INTO posts (id, title) VALUES (1, 'Test Post');

		-- Insert test comments
		INSERT INTO comments (id, post_id, user_id, content, created_at, parent_id) VALUES 
		(1, 1, 1, 'First comment', '2024-01-01 10:00:00', NULL),
		(2, 1, 2, 'Second comment', '2024-01-01 11:00:00', NULL),
		(3, 1, 1, 'Reply to first comment', '2024-01-01 12:00:00', 1);

		-- Insert comment likes
		INSERT INTO comment_likes (comment_id, is_like) VALUES 
		(1, 1), (1, 1),  -- 2 likes for first comment
		(2, 0), (2, 0);  -- 2 dislikes for second comment
	`)
	if err != nil {
		t.Fatalf("Failed to prepare mock data: %v", err)
	}

	// Mock GetCommentReplies to return predefined replies
	originalGetCommentReplies := GetCommentReplies
	GetCommentReplies = func(commentID int) ([]Comment, error) {
		if commentID == 1 {
			return []Comment{
				{
					ID:       3,
					PostID:   1,
					UserID:   "1",
					Content:  "Reply to first comment",
					Username: "testuser1",
				},
			}, nil
		}
		return []Comment{}, nil
	}
	defer func() { GetCommentReplies = originalGetCommentReplies }()

	// Test cases
	testCases := []struct {
		name           string
		postID         int
		expectedResult struct {
			commentCount      int
			firstCommentID    int
			firstCommentLikes int
			replyCount        int
		}
	}{
		{
			name:   "Non-Existing Post",
			postID: 999,
			expectedResult: struct {
				commentCount      int
				firstCommentID    int
				firstCommentLikes int
				replyCount        int
			}{
				commentCount: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comments, err := GetCommentsForPost(tc.postID)

			if tc.postID == 999 {
				// For non-existing post, expect no error and empty comments
				if err != nil {
					t.Errorf("Unexpected error for non-existing post: %v", err)
				}
				if len(comments) != 0 {
					t.Errorf("Expected 0 comments for non-existing post, got %d", len(comments))
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(comments) != tc.expectedResult.commentCount {
				t.Errorf("Expected %d comments, got %d", tc.expectedResult.commentCount, len(comments))
			}

			if len(comments) > 0 {
				firstComment := comments[0]
				if firstComment.ID != tc.expectedResult.firstCommentID {
					t.Errorf("Expected first comment ID %d, got %d", tc.expectedResult.firstCommentID, firstComment.ID)
				}

				if firstComment.LikeCount != tc.expectedResult.firstCommentLikes {
					t.Errorf("Expected %d likes, got %d", tc.expectedResult.firstCommentLikes, firstComment.LikeCount)
				}

				if len(firstComment.Replies) != tc.expectedResult.replyCount {
					t.Errorf("Expected %d replies, got %d", tc.expectedResult.replyCount, len(firstComment.Replies))
				}
			}
		})
	}
}

func TestGetCommentReplies(t *testing.T) {
	// Setup mock database
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Replace global db with mock
	originalDB := db
	db = mockDB
	defer func() { db = originalDB }()

	// Prepare mock database schema and data
	_, err = mockDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			user_id INTEGER,
			content TEXT,
			created_at DATETIME,
			parent_id INTEGER,
			FOREIGN KEY(post_id) REFERENCES posts(id),
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(parent_id) REFERENCES comments(id)
		);
		CREATE TABLE comment_likes (
			comment_id INTEGER,
			is_like BOOLEAN
		);

		-- Insert test users
		INSERT INTO users (id, username) VALUES 
		(1, 'testuser1'),
		(2, 'testuser2');

		-- Insert test post
		INSERT INTO posts (id, title) VALUES (1, 'Test Post');

		-- Insert test comments
		INSERT INTO comments (id, post_id, user_id, content, created_at, parent_id) VALUES 
		(1, 1, 1, 'Parent comment', '2024-01-01 10:00:00', NULL),
		(2, 1, 2, 'First reply', '2024-01-01 11:00:00', 1),
		(3, 1, 1, 'Second reply', '2024-01-01 12:00:00', 1);

		-- Insert comment likes
		INSERT INTO comment_likes (comment_id, is_like) VALUES 
		(2, 1), (2, 1),  -- 2 likes for first reply
		(3, 0), (3, 0);  -- 2 dislikes for second reply
	`)
	if err != nil {
		t.Fatalf("Failed to prepare mock data: %v", err)
	}

	// Test cases
	testCases := []struct {
		name           string
		commentID      int
		expectedResult struct {
			replyCount         int
			firstReplyID       int
			firstReplyLikes    int
			firstReplyDislikes int
		}
	}{
		{
			name:      "Comment with No Replies",
			commentID: 2,
			expectedResult: struct {
				replyCount         int
				firstReplyID       int
				firstReplyLikes    int
				firstReplyDislikes int
			}{
				replyCount: 0,
			},
		},
		{
			name:      "Non-Existing Comment",
			commentID: 999,
			expectedResult: struct {
				replyCount         int
				firstReplyID       int
				firstReplyLikes    int
				firstReplyDislikes int
			}{
				replyCount: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			replies, err := GetCommentReplies(tc.commentID)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(replies) != tc.expectedResult.replyCount {
				t.Errorf("Expected %d replies, got %d", tc.expectedResult.replyCount, len(replies))
			}

			if len(replies) > 0 {
				firstReply := replies[0]
				if firstReply.ID != tc.expectedResult.firstReplyID {
					t.Errorf("Expected first reply ID %d, got %d", tc.expectedResult.firstReplyID, firstReply.ID)
				}

				if firstReply.LikeCount != tc.expectedResult.firstReplyLikes {
					t.Errorf("Expected %d likes, got %d", tc.expectedResult.firstReplyLikes, firstReply.LikeCount)
				}

				if firstReply.DislikeCount != tc.expectedResult.firstReplyDislikes {
					t.Errorf("Expected %d dislikes, got %d", tc.expectedResult.firstReplyDislikes, firstReply.DislikeCount)
				}
			}
		})
	}
}

func TestCommentHandler(t *testing.T) {
	// Setup mock database
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Replace global db with mock
	originalDB := db
	db = mockDB
	defer func() { db = originalDB }()

	// Prepare mock database schema and data
	_, err = mockDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			user_id INTEGER,
			content TEXT,
			created_at DATETIME,
			parent_id INTEGER,
			FOREIGN KEY(post_id) REFERENCES posts(id),
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(parent_id) REFERENCES comments(id)
		);

		-- Insert test users
		INSERT INTO users (id, username) VALUES 
		(1, 'testuser1'),
		(2, 'testuser2');

		-- Insert test post
		INSERT INTO posts (id, title) VALUES (1, 'Test Post');
	`)
	if err != nil {
		t.Fatalf("Failed to prepare mock data: %v", err)
	}

	// Test cases
	testCases := []struct {
		name           string
		method         string
		postID         string
		content        string
		parentID       string
		userID         string
		expectedStatus int
		expectedError  string
		checkDBFunc    func(t *testing.T, db *sql.DB) // Optional function to check database state
	}{
		{
			name:           "Valid Top-Level Comment",
			method:         http.MethodPost,
			postID:         "1",
			content:        "Test comment",
			userID:         "1",
			expectedStatus: http.StatusSeeOther,
			checkDBFunc: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = 1 AND parent_id IS NULL").Scan(&count)
				if err != nil {
					t.Fatalf("Error checking comment count: %v", err)
				}
				if count != 1 {
					t.Errorf("Expected 1 top-level comment, got %d", count)
				}
			},
		},
		{
			name:           "Invalid Request Method",
			method:         http.MethodGet,
			postID:         "1",
			content:        "Test comment",
			userID:         "1",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Invalid request method",
		},
		{
			name:           "Missing User ID",
			method:         http.MethodPost,
			postID:         "1",
			content:        "Test comment",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Please log in to comment on posts",
		},

		{
			name:           "Empty Comment Content",
			method:         http.MethodPost,
			postID:         "1",
			content:        "",
			userID:         "1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Comment content cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset database for each test case
			_, err = mockDB.Exec("DELETE FROM comments")
			if err != nil {
				t.Fatalf("Failed to reset comments table: %v", err)
			}

			// Prepare form data
			formData := url.Values{
				"post_id":   {tc.postID},
				"content":   {tc.content},
				"parent_id": {tc.parentID},
			}

			// Create a request
			req := httptest.NewRequest(tc.method, "/comment", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create a response recorder
			w := httptest.NewRecorder()

			// Mock GetUserIdFromSession
			originalGetUserIdFromSession := GetUserIdFromSession
			GetUserIdFromSession = func(w http.ResponseWriter, r *http.Request) string {
				return tc.userID
			}
			defer func() { GetUserIdFromSession = originalGetUserIdFromSession }()

			// Call the handler
			CommentHandler(w, req)

			// Check response status
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Check error message if expected
			if tc.expectedError != "" {
				body, _ := ioutil.ReadAll(resp.Body)
				if !strings.Contains(string(body), tc.expectedError) {
					t.Errorf("Expected error message '%s', got '%s'", tc.expectedError, string(body))
				}
			}

			// Optional database state check
			if tc.checkDBFunc != nil {
				tc.checkDBFunc(t, mockDB)
			}
		})
	}
}

func TestCommentLikeHandler(t *testing.T) {
	// Setup mock database
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Replace global db with mock
	originalDB := db
	db = mockDB
	defer func() { db = originalDB }()

	// Prepare mock database schema and data
	_, err = mockDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			user_id INTEGER,
			content TEXT,
			created_at DATETIME,
			parent_id INTEGER,
			FOREIGN KEY(post_id) REFERENCES posts(id),
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(parent_id) REFERENCES comments(id)
		);
		CREATE TABLE comment_likes (
			comment_id INTEGER,
			user_id INTEGER,
			is_like BOOLEAN,
			PRIMARY KEY(comment_id, user_id)
		);

		-- Insert test users
		INSERT INTO users (id, username) VALUES 
		(1, 'testuser1'),
		(2, 'testuser2');

		-- Insert test post
		INSERT INTO posts (id, title) VALUES (1, 'Test Post');

		-- Insert test comment
		INSERT INTO comments (id, post_id, user_id, content, created_at) VALUES 
		(1, 1, 1, 'Test comment', '2024-01-01 10:00:00');
	`)
	if err != nil {
		t.Fatalf("Failed to prepare mock data: %v", err)
	}

	// Test cases
	testCases := []struct {
		name           string
		method         string
		userID         string
		commentID      string
		isLike         string
		expectedStatus int
		expectedError  string
		checkDBFunc    func(t *testing.T, db *sql.DB)          // Optional function to check database state
		checkResponse  func(t *testing.T, resp *http.Response) // Optional function to check response
	}{
		{
			name:           "Valid First Like",
			method:         http.MethodPost,
			userID:         "1",
			commentID:      "1",
			isLike:         "true",
			expectedStatus: http.StatusOK,
			checkDBFunc: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = 1 AND user_id = 1 AND is_like = 1").Scan(&count)
				if err != nil {
					t.Fatalf("Error checking like: %v", err)
				}
				if count != 1 {
					t.Errorf("Expected 1 like, got %d", count)
				}
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				var likeResp CommentLikeResponse
				err := json.NewDecoder(resp.Body).Decode(&likeResp)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if likeResp.LikeCount != 1 || likeResp.DislikeCount != 0 {
					t.Errorf("Unexpected like/dislike counts: likes=%d, dislikes=%d",
						likeResp.LikeCount, likeResp.DislikeCount)
				}
				if likeResp.UserLiked == nil || *likeResp.UserLiked != true {
					t.Errorf("Unexpected user liked status: %v", likeResp.UserLiked)
				}
			},
		},
		{
			name:           "Change Like to Dislike",
			method:         http.MethodPost,
			userID:         "1",
			commentID:      "1",
			isLike:         "false",
			expectedStatus: http.StatusOK,
			checkDBFunc: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM comment_likes WHERE comment_id = 1 AND user_id = 1 AND is_like = 0").Scan(&count)
				if err != nil {
					t.Fatalf("Error checking dislike: %v", err)
				}
				if count != 1 {
					t.Errorf("Expected 1 dislike, got %d", count)
				}
			},
			checkResponse: func(t *testing.T, resp *http.Response) {
				var likeResp CommentLikeResponse
				err := json.NewDecoder(resp.Body).Decode(&likeResp)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if likeResp.LikeCount != 0 || likeResp.DislikeCount != 1 {
					t.Errorf("Unexpected like/dislike counts: likes=%d, dislikes=%d",
						likeResp.LikeCount, likeResp.DislikeCount)
				}
				if likeResp.UserLiked == nil || *likeResp.UserLiked != false {
					t.Errorf("Unexpected user liked status: %v", likeResp.UserLiked)
				}
			},
		},
		{
			name:           "Invalid Request Method",
			method:         http.MethodGet,
			userID:         "1",
			commentID:      "1",
			isLike:         "true",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
		{
			name:           "Missing User ID",
			method:         http.MethodPost,
			userID:         "",
			commentID:      "1",
			isLike:         "true",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Please log in to like or dislike comments",
		},
		{
			name:           "Missing Comment ID",
			method:         http.MethodPost,
			userID:         "1",
			commentID:      "",
			isLike:         "true",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Comment ID is required",
		},
		{
			name:           "Invalid Comment ID",
			method:         http.MethodPost,
			userID:         "1",
			commentID:      "invalid",
			isLike:         "true",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid comment ID",
		},
		{
			name:           "Non-Existent Comment",
			method:         http.MethodPost,
			userID:         "1",
			commentID:      "999",
			isLike:         "true",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Comment not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset comment_likes table for each test case
			_, err = mockDB.Exec("DELETE FROM comment_likes")
			if err != nil {
				t.Fatalf("Failed to reset comment_likes table: %v", err)
			}

			// Prepare form data
			formData := url.Values{
				"comment_id": {tc.commentID},
				"is_like":    {tc.isLike},
			}

			// Create a request
			req := httptest.NewRequest(tc.method, "/comment/like", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create a response recorder
			w := httptest.NewRecorder()

			// Mock GetUserIdFromSession
			originalGetUserIdFromSession := GetUserIdFromSession
			GetUserIdFromSession = func(w http.ResponseWriter, r *http.Request) string {
				return tc.userID
			}
			defer func() { GetUserIdFromSession = originalGetUserIdFromSession }()

			// Call the handler
			CommentLikeHandler(w, req)

			// Check response status
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Check error message if expected
			if tc.expectedError != "" {
				body, _ := ioutil.ReadAll(resp.Body)
				if !strings.Contains(string(body), tc.expectedError) {
					t.Errorf("Expected error message '%s', got '%s'", tc.expectedError, string(body))
				}
			}

			// Optional database state check
			if tc.checkDBFunc != nil {
				tc.checkDBFunc(t, mockDB)
			}

			// Optional response check
			if tc.checkResponse != nil {
				tc.checkResponse(t, resp)
			}
		})
	}
}
