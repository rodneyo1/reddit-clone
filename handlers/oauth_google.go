package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2v2 "google.golang.org/api/oauth2/v2"
)

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  string
)

// InitGoogleOAuth initializes the Google OAuth configuration
func InitGoogleOAuth() {
	// Get environment variables
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	// Verify required environment variables
	if clientID == "" || clientSecret == "" {
		fmt.Println("Error: GOOGLE_CLIENT_ID and/or GOOGLE_CLIENT_SECRET environment variables are not set")
		return
	}

	googleOauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8081/auth/google/callback",
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// addSecurityHeaders adds necessary security headers to the response
func addSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy",
		"default-src 'self' https://*.google.com https://accounts.google.com; "+
			"script-src 'self' 'unsafe-inline' https://*.google.com https://accounts.google.com; "+
			"frame-src https://*.google.com https://accounts.google.com; "+
			"img-src 'self' https: data:; "+
			"style-src 'self' 'unsafe-inline' https://*.google.com https://accounts.google.com https://cdnjs.cloudflare.com; "+
			"font-src 'self' https://cdnjs.cloudflare.com; "+
			"connect-src 'self' https://*.google.com https://accounts.google.com")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

// HandleGoogleLogin handles the Google OAuth login/registration flow
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	addSecurityHeaders(w)

	// Check if this is a registration or login flow
	isRegistration := false
	if referer := r.Header.Get("Referer"); referer != "" {
		isRegistration = strings.Contains(referer, "/register")
	}

	//starting the goolgle auth process
	if googleOauthConfig == nil {
		RenderError(w, r, "Internal server. Please try again later.", http.StatusInternalServerError)
		log.Println("Error: Google OAuth configuration is nil")
		return
	}

	// Generate a simple state that includes the flow type
	stateUUID := uuid.New().String()
	state := fmt.Sprintf("%s:%v", stateUUID, isRegistration)

	// Store the state
	oauthStateString = state

	// Generate OAuth URL with the state
	url := googleOauthConfig.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback handles the Google OAuth callback
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	addSecurityHeaders(w)

	// Parse form data to ensure we can access all parameters
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
		// Render a generic error message to the user
		RenderError(w, r, "An error occurred during authentication. Please try again.", http.StatusUnauthorized)
		return
	}

	// Verify state parameter
	stateParam := r.FormValue("state")

	if stateParam != oauthStateString {
		fmt.Printf("State mismatch! >>>>>>>>> Recieved: %s, Expected: %s\n", stateParam, oauthStateString)
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
		fmt.Println("No code received from Google")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//attempt to exchange the code for a token
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get user info from Google
	client := googleOauthConfig.Client(r.Context(), token)
	service, err := oauth2v2.New(client)
	if err != nil {
		fmt.Printf("Failed to create OAuth2 service: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	userInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		fmt.Printf("Failed to get user info: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Check if user exists (by email or Google ID)
	var userID string
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", userInfo.Email, userInfo.Id).Scan(&userID)
	if err != nil {
		if isRegister {
			// Check if email exists with traditional login
			var traditionalUserID string
			err = db.QueryRow("SELECT id FROM users WHERE email = ?", userInfo.Email).Scan(&traditionalUserID)
			if err == nil {
				// Email exists with traditional login
				RenderError(w, r, "An account with this email already exists. Please login with your password or use 'Forgot Password'.", http.StatusConflict)
				return
			}

			// Create new user for registration
			userID = uuid.New().String()
			_, err = db.Exec(`
	            INSERT INTO users (id, email, username, google_id, avatar_url)
	            VALUES (?, ?, ?, ?, ?)`,
				userID, userInfo.Email, userInfo.Name, userInfo.Id, userInfo.Picture)
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
			// Check if the account has a Google ID
			var existingGoogleID string
			err = db.QueryRow("SELECT google_id FROM users WHERE id = ?", userID).Scan(&existingGoogleID)
			if err == nil && existingGoogleID != "" {
				RenderError(w, r, "An account with this Google ID already exists. Please login instead.", http.StatusConflict)
				return
			}

			// Update existing account with Google ID
			_, err = db.Exec("UPDATE users SET google_id = ? WHERE id = ?", userInfo.Id, userID)
			if err != nil {
				RenderError(w, r, "Failed to link Google account. Please try again.", http.StatusInternalServerError)
				return
			}
			w.Header().Set("X-Auth-Message", fmt.Sprintf("Your Google account has been linked successfully! Welcome, %s!", userInfo.Name))
		} else {
			w.Header().Set("X-Auth-Message", fmt.Sprintf("Welcome back, %s!", userInfo.Name))
		}
	}

	// Store tokens
	expiresAt := time.Now().Add(time.Second * time.Duration(token.Expiry.Unix()-time.Now().Unix()))

	_, err = db.Exec(`
	INSERT OR REPLACE INTO google_auth (user_id, access_token, refresh_token, expires_at)
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
