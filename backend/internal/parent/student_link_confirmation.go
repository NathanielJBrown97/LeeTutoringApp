// backend/internal/parent/confirm_link_students.go

package parent

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
)

// ConfirmLinkStudentsHandler handles the confirmation of linking students
func (a *App) ConfirmLinkStudentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the session from the App's Store
	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	// Extract the user ID from the session
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		http.Error(w, "UserID not found in session", http.StatusUnauthorized)
		return
	}

	// Parse the form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Reference to the parent's document using the userID from the session
	parentDocRef := firestoreClient.Collection("parents").Doc(userID)

	// Loop through the form data and link confirmed students
	for key, values := range r.PostForm {
		if strings.HasPrefix(key, "confirm_") && len(values) > 0 && values[0] == "yes" {
			studentID := strings.TrimPrefix(key, "confirm_")
			_, err = parentDocRef.Update(context.Background(), []firestore.Update{
				{
					Path:  "associated_students",
					Value: firestore.ArrayUnion(studentID),
				},
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to link student ID %s: %v", studentID, err), http.StatusInternalServerError)
				return
			}
		}
	}

	// Redirect to the parent dashboard after linking
	http.Redirect(w, r, "/parentdashboard", http.StatusSeeOther)
}
