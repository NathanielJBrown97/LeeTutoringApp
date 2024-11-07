// backend/internal/auth/handler.go

package auth

import (
	"encoding/json"
	"log"
	"net/http"
)

// AuthStatus represents the structure of the authentication status response
type AuthStatus struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"user_id,omitempty"`
	Email         string `json:"email,omitempty"`
}

/// backend/internal/auth/handler.go

func (a *App) StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("StatusHandler invoked")

	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		log.Printf("Failed to retrieve session: %v", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	userID, ok := session.Values["user_id"].(string)
	email, emailOK := session.Values["user_email"].(string)

	if !ok || !emailOK || userID == "" || email == "" {
		log.Println("User is not authenticated.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthStatus{Authenticated: false})
		return
	}

	log.Printf("User authenticated: %s (%s)", userID, email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthStatus{
		Authenticated: true,
		UserID:        userID,
		Email:         email,
	})
}
