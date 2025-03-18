package handlers

import (
	"net/http"
)

// HandleDatabaseError handles database-related errors
func HandleDatabaseError(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		RenderError(w, r, "Database Error", http.StatusInternalServerError)
		return
	}
}
