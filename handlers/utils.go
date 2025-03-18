package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// ErrorData represents the data passed to the error template
type ErrorData struct {
	StatusCode   int
	ErrorMessage string
	HelpMessage  string
	IsLoggedIn   bool
}

// Common error messages and help text
var (
	ErrorMessages = map[string]ErrorData{
		// Authentication errors
		"invalid_credentials": {
			StatusCode:   http.StatusUnauthorized,
			ErrorMessage: "Invalid email or password",
			HelpMessage:  "Please check your email and password and try again. If you've forgotten your password, you can reset it.",
		},
		"unauthorized": {
			StatusCode:   http.StatusUnauthorized,
			ErrorMessage: "You must be logged in to perform this action",
			HelpMessage:  "Please log in to your account to continue.",
		},

		// Registration errors
		"invalid_email": {
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: "Invalid email format",
			HelpMessage:  "Please enter a valid email address (e.g., user@example.com).",
		},
		"email_exists": {
			StatusCode:   http.StatusConflict,
			ErrorMessage: "Email already registered",
			HelpMessage:  "This email is already registered. Please use a different email or try logging in.",
		},
		"password_too_short": {
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: "Password is too short",
			HelpMessage:  "Password must be at least 6 characters long. Use a mix of letters, numbers, and symbols for better security.",
		},
		"passwords_dont_match": {
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: "Passwords do not match",
			HelpMessage:  "The passwords you entered don't match. Please try again.",
		},

		// Input validation errors
		"invalid_input": {
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: "Invalid input provided",
			HelpMessage:  "Please check your input and try again. All required fields must be filled out.",
		},
		"missing_fields": {
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: "Required fields are missing",
			HelpMessage:  "Please fill out all required fields marked with an asterisk (*).",
		},

		// Resource errors
		"post_not_found": {
			StatusCode:   http.StatusNotFound,
			ErrorMessage: "Post not found",
			HelpMessage:  "The post you're looking for might have been deleted or never existed.",
		},
		"comment_not_found": {
			StatusCode:   http.StatusNotFound,
			ErrorMessage: "Comment not found",
			HelpMessage:  "The comment you're looking for might have been deleted or never existed.",
		},

		// Server errors
		"database_error": {
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Database error occurred",
			HelpMessage:  "We're experiencing technical difficulties. Please try again later.",
		},
		"server_error": {
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
			HelpMessage:  "Something went wrong on our end. Please try again later.",
		},

		// Permission errors
		"forbidden": {
			StatusCode:   http.StatusForbidden,
			ErrorMessage: "Access denied",
			HelpMessage:  "You don't have permission to perform this action.",
		},
		"not_owner": {
			StatusCode:   http.StatusForbidden,
			ErrorMessage: "Not the owner",
			HelpMessage:  "You can only modify content that you created.",
		},
	}
)

// RenderErrorFunc is the type for rendering error responses
type RenderErrorFunc func(w http.ResponseWriter, r *http.Request, errorKey string, statusCode int)

// RenderError renders the error template with the given error message and status code
func renderError(w http.ResponseWriter, r *http.Request, errorKey string, statusCode int) {
	// Get user's login status
	userID := GetUserIdFromSession(w, r)
	isLoggedIn := userID != ""

	// Get error data from the map, or use default if not found
	errorData, exists := ErrorMessages[errorKey]
	if !exists {
		errorData = ErrorData{
			StatusCode:   statusCode,
			ErrorMessage: errorKey,
			HelpMessage:  "Please try again or contact support if the problem persists.",
		}
	}
	errorData.IsLoggedIn = isLoggedIn

	// Parse and execute the error template
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		log.Printf("Error parsing error template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(errorData.StatusCode)
	if err := tmpl.Execute(w, errorData); err != nil {
		log.Printf("Error executing error template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// RenderError is a function variable that can be mocked in tests
var RenderError RenderErrorFunc = renderError
