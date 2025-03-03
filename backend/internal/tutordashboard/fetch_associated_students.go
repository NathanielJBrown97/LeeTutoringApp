// backend/internal/tutordashboard/fetch_associated_students.go

package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// AssociatedStudent represents a summary of an associated student's data.
type AssociatedStudent struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// FetchAssociatedStudentsHandler returns an HTTP handler that fetches all associated students
// for a given tutor. It expects the tutor's user ID to be provided as a query parameter "tutorUserID".
// The handler queries the "Associated Students" subcollection under the tutor's document and returns
// an array of AssociatedStudent objects.
func FetchAssociatedStudentsHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// For testing purposes, we extract tutorUserID from query parameters.
		tutorUserID := r.URL.Query().Get("tutorUserID")
		if tutorUserID == "" {
			http.Error(w, "Missing tutorUserID", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		// Reference the "Associated Students" subcollection for the given tutor.
		iter := client.Collection("tutors").Doc(tutorUserID).Collection("Associated Students").Documents(ctx)
		defer iter.Stop()

		var students []AssociatedStudent
		for {
			doc, err := iter.Next()
			if err != nil {
				// When the iterator is finished, break out of the loop.
				break
			}

			data := doc.Data()
			// Optionally, read the student's name from the data.
			studentName, _ := data["name"].(string)
			students = append(students, AssociatedStudent{
				ID:   doc.Ref.ID,
				Name: studentName,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(students)
	}
}
