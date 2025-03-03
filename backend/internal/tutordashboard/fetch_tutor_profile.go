// backend/internal/tutordashboard/fetch_tutor_profile.go

package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// TutorProfile holds the basic profile information for a tutor.
type TutorProfile struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	UserID  string `json:"user_id"`
}

// FetchTutorProfileHandler returns an HTTP handler that fetches a tutor's profile.
// For testing purposes, it accepts a query parameter "tutorUserID".
// In production, you would extract the tutor's ID from the authenticated context.
func FetchTutorProfileHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// For testing, fetch tutorUserID from query parameters.
		tutorUserID := r.URL.Query().Get("tutorUserID")
		if tutorUserID == "" {
			http.Error(w, "Missing tutorUserID", http.StatusBadRequest)
			return
		}

		// Retrieve the tutor document from the "tutors" collection.
		doc, err := client.Collection("tutors").Doc(tutorUserID).Get(r.Context())
		if err != nil {
			http.Error(w, "Failed to fetch tutor profile", http.StatusInternalServerError)
			return
		}

		data := doc.Data()

		// Build the tutor profile from the document data.
		profile := TutorProfile{
			UserID: tutorUserID,
		}
		if email, ok := data["email"].(string); ok {
			profile.Email = email
		}
		if name, ok := data["name"].(string); ok {
			profile.Name = name
		}
		if picture, ok := data["picture"].(string); ok {
			profile.Picture = picture
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}
}
