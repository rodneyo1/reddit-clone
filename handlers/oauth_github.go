package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	githubOauthConfig *oauth2.Config
	githubStateString string
)

// InitGithubOAuth initializes the GitHub OAuth configuration
func InitGithubOAuth() {
	// Get environment variables
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	// Verify required environment variables
	if clientID == "" || clientSecret == "" {
		fmt.Println("Error: GITHUB_CLIENT_ID and/or GITHUB_CLIENT_SECRET environment variables are not set")
		return
	}

	githubOauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8081/auth/github/callback",
		Scopes: []string{
			"user:email",
			"read:user",
		},
		Endpoint: github.Endpoint,
	}
}

// HandleGithubLogin handles the GitHub OAuth login/registration flow
func HandleGithubLogin(w http.ResponseWriter, r *http.Request) {
	addSecurityHeaders(w)

	// Check if this is a registration or login flow
	isRegistration := false
	if referer := r.Header.Get("Referer"); referer != "" {
		isRegistration = strings.Contains(referer, "/register")
	}

	if githubOauthConfig == nil {
		RenderError(w, r, "Internal server error. Please try again later.", http.StatusInternalServerError)
		log.Println("Error: GitHub OAuth configuration is nil")
		return
	}

	// Generate a simple state that includes the flow type
	stateUUID := uuid.New().String()
	state := fmt.Sprintf("%s:%v", stateUUID, isRegistration)

	// Store the state
	githubStateString = state

	// Generate OAuth URL with the state
	url := githubOauthConfig.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GithubUser represents the GitHub user data structure
type GithubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// HandleGithubCallback handles the GitHub OAuth callback
func HandleGithubCallback(w http.ResponseWriter, r *http.Request) {
	addSecurityHeaders(w)

	// Parse form data
	if err := r.ParseForm(); err != nil {
		fmt.Printf("Error parsing form data: %v\n", err)
		RenderError(w, r, "Error processing callback data", http.StatusBadRequest)
		return
	}

	// Set SameSite attribute for all cookies
	w.Header().Set("Set-Cookie", "SameSite=Lax")

	// Check if there's an error in the callback
	if errMsg := r.FormValue("error"); errMsg != "" {
		fmt.Printf("OAuth Error: %s\n", errMsg)
		RenderError(w, r, "An error occurred during authentication. Please try again.", http.StatusUnauthorized)
		return
	}

	// Verify state parameter
	stateParam := r.FormValue("state")
	if stateParam != githubStateString {
		fmt.Printf("github State mismatch! Received: %s, Expected: %s\n", stateParam, githubStateString)
		RenderError(w, r, "Invalid authentication state. Please try again.", http.StatusBadRequest)
		return
	}

	// Parse state parameter (format: "uuid:isRegister")
	stateParts := strings.Split(stateParam, ":")
	if len(stateParts) != 2 {
		fmt.Printf("Invalid state format: %s\n", stateParam)
		RenderError(w, r, "Invalid authentication state.", http.StatusBadRequest)
		return
	}

	// Extract flow type
	isRegister := stateParts[1] == "true"

	// Exchange auth code for token
	code := r.FormValue("code")
	if code == "" {
		fmt.Println("No code received from GitHub")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Exchange the code for a token
	token, err := githubOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		fmt.Printf("Failed to exchange token: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get user info from GitHub
	client := githubOauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		fmt.Printf("Failed to get user info: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var githubUser GithubUser
	if err := json.Unmarshal(body, &githubUser); err != nil {
		fmt.Printf("Failed to parse user data: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// If email is not public, fetch it from the email endpoint
	if githubUser.Email == "" {
		emailsResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailsResp.Body.Close()
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err := json.NewDecoder(emailsResp.Body).Decode(&emails); err == nil {
				for _, email := range emails {
					if email.Primary && email.Verified {
						githubUser.Email = email.Email
						break
					}
				}
			}
		}
	}

	// Check if user exists by email
	var userID string
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", githubUser.Email).Scan(&userID)
	if err != nil {
		// No user found with this email
		if isRegister {
			// Check if email exists with traditional login
			var traditionalUserID string
			err = db.QueryRow("SELECT id FROM users WHERE email = ?", githubUser.Email).Scan(&traditionalUserID)
			if err == nil {
				// Email exists with traditional login
				RenderError(w, r, "An account with this email already exists. Please login with your password or use 'Forgot Password'.", http.StatusConflict)
				return
			}

			// Create new user for registration
			userID = uuid.New().String()
			// Use Login (username) as fallback if Name is empty
			username := githubUser.Name
			if username == "" {
				username = githubUser.Login
			}

			_, err = db.Exec(`
			INSERT INTO users (id, email, username, github_id, avatar_url)
			VALUES (?, ?, ?, ?, ?)`,
				userID, githubUser.Email, username, githubUser.ID, githubUser.AvatarURL)
			if err != nil {
				RenderError(w, r, "Failed to complete registration. Please try again.", http.StatusInternalServerError)
				return
			}
		} else {
			// User doesn't exist and trying to login
			RenderError(w, r, "No account found with this email. Please register first.", http.StatusUnauthorized)
			return
		}
	} else {
		// User exists
		if isRegister {
			// Check if the account has a GitHub ID
			var existingGithubID string
			err = db.QueryRow("SELECT github_id FROM users WHERE id = ?", userID).Scan(&existingGithubID)
			if err == nil && existingGithubID != "" {
				RenderError(w, r, "An account with this GitHub ID already exists. Please login instead.", http.StatusConflict)
				return
			}

			// Get current username
			var currentUsername string
			err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&currentUsername)
			if err == nil && currentUsername == "" {
				// Username is empty, update it with GitHub username
				username := githubUser.Name
				if username == "" {
					username = githubUser.Login
				}
				_, err = db.Exec("UPDATE users SET github_id = ?, username = ? WHERE id = ?", githubUser.ID, username, userID)
			} else {
				// Just update GitHub ID
				_, err = db.Exec("UPDATE users SET github_id = ? WHERE id = ?", githubUser.ID, userID)
			}
			if err != nil {
				RenderError(w, r, "Failed to link GitHub account. Please try again.", http.StatusInternalServerError)
				return
			}
			w.Header().Set("X-Auth-Message", fmt.Sprintf("Your GitHub account has been linked successfully! Welcome, %s!", githubUser.Name))
		} else {
			w.Header().Set("X-Auth-Message", fmt.Sprintf("Welcome back, %s!", githubUser.Name))
		}
	}

	// Store tokens
	expiresAt := time.Now().Add(time.Hour * 24) // GitHub tokens typically don't have an expiry
	_, err = db.Exec(`
		INSERT OR REPLACE INTO github_auth (user_id, access_token, refresh_token, expires_at)
		VALUES (?, ?, ?, ?)`,
		userID, token.AccessToken, token.RefreshToken, expiresAt)
	if err != nil {
		RenderError(w, r, "Failed to complete authentication. Please try signing in again.", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID := uuid.New().String()
	_, err = db.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, userID)
	if err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
		RenderError(w, r, "Internal server error. Please try signing in again.", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	cookieExpiry := time.Now().Add(24 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  cookieExpiry,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to home page with appropriate status
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
