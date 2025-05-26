package handlers

import (
	"net/http"
	"time"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Always remove the cookie (even if session doesn't exist)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Try deleting from DB if the session exists
	if sessionCookie, err := r.Cookie("session_id"); err == nil {
		_, err = db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionCookie.Value)
		if err != nil {
			// Log the error but don't block logout
			http.Error(w, "Error deleting session", http.StatusInternalServerError)
			return
		}
	}

	// ✅ Always return JSON success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Logged out successfully"}`))
}
