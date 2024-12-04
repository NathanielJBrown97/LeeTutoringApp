// backend/internal/microsoftauth/picture.go

package microsoftauth

import (
	"context"
	"io/ioutil"
	"net/http"
)

func (a *App) ProfilePictureHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Retrieve the user's access token from Firestore
	docRef := a.FirestoreClient.Collection("parents").Doc(userID)
	doc, err := docRef.Get(context.Background())
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := doc.Data()
	accessToken, ok := data["access_token"].(string)
	if !ok || accessToken == "" {
		http.Error(w, "Access token not found", http.StatusUnauthorized)
		return
	}

	// Fetch the user's profile picture
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/photo/$value", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get profile picture", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Set the appropriate content type
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))

	// Stream the picture directly to the response
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read picture data", http.StatusInternalServerError)
		return
	}
}
