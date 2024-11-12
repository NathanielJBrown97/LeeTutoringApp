// backend/internal/auth/handler.go

package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/NathanielJBrown97/LeeTutoringApp/internal/middleware"
)

// AuthStatus represents the structure of the authentication status response
type AuthStatus struct {
	Authenticated bool   `json:"authenticated"`
	UserID        string `json:"user_id,omitempty"`
	Email         string `json:"email,omitempty"`
}

func (a *App) StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("StatusHandler invoked")

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("User is not authenticated.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthStatus{Authenticated: false})
		return
	}

	userID, _ := claims["user_id"].(string)
	email, _ := claims["email"].(string)

	log.Printf("User authenticated: %s (%s)", userID, email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthStatus{
		Authenticated: true,
		UserID:        userID,
		Email:         email,
	})
}
