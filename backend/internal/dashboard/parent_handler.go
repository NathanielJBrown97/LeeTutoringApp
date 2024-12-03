// backend/internal/dashboard/parent_handler.go

package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type ParentData struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// ParentHandler serves the parent data as JSON
func (a *App) ParentHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user ID from the JWT token
	userID, email := a.getParentCredentials(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch parent data from Firestore
	ctx := context.Background()
	parentDocRef := a.FirestoreClient.Collection("parents").Doc(userID)
	docSnap, err := parentDocRef.Get(ctx)
	if err != nil {
		log.Printf("Error fetching parent data: %v", err)
		http.Error(w, "Failed to fetch parent data", http.StatusInternalServerError)
		return
	}

	data := docSnap.Data()
	parentData := ParentData{
		UserID: userID,
		Email:  email,
	}

	// Safely extract name
	if name, ok := data["name"].(string); ok {
		parentData.Name = name
	} else {
		parentData.Name = ""
	}

	// Safely extract picture URL
	if picture, ok := data["picture"].(string); ok {
		parentData.Picture = picture
	} else {
		parentData.Picture = ""
	}

	// Return the parent data as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(parentData); err != nil {
		log.Printf("Error encoding parent data to JSON: %v", err)
		http.Error(w, "Failed to encode parent data", http.StatusInternalServerError)
	}
}