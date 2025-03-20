// edit_personal_details.go
package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// EditPersonalDetailsRequest defines the expected payload for editing a student's personal details.
type EditPersonalDetailsRequest struct {
	FirebaseID     string `json:"firebase_id"`     // The student's Firebase ID.
	Name           string `json:"name"`            // The student's name.
	Accommodations string `json:"accommodations"`  // Any accommodations.
	Grade          string `json:"grade"`           // The student's grade.
	HighSchool     string `json:"high_school"`     // The high school the student attends.
	ParentEmail    string `json:"parent_email"`    // Parent email address.
	StudentEmail   string `json:"student_email"`   // Student email address.
	Interests      string `json:"interests"`       // Student interests.
	ParentNumber   string `json:"parent_number"`   // Parent contact number.
	StudentNumber  string `json:"student_number"`  // Student contact number.
}

// EditPersonalDetailsHandler returns an HTTP handler function that processes an edit personal details request.
func EditPersonalDetailsHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req EditPersonalDetailsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" {
			http.Error(w, "Missing required field: firebase_id", http.StatusBadRequest)
			return
		}

		// Build the update data using field paths to update the nested "personal" subdocument.
		updates := []firestore.Update{
			{Path: "personal.name", Value: req.Name},
			{Path: "personal.accommodations", Value: req.Accommodations},
			{Path: "personal.grade", Value: req.Grade},
			{Path: "personal.high_school", Value: req.HighSchool},
			{Path: "personal.parent_email", Value: req.ParentEmail},
			{Path: "personal.student_email", Value: req.StudentEmail},
			{Path: "personal.interests", Value: req.Interests},
			{Path: "personal.parent_number", Value: req.ParentNumber},
			{Path: "personal.student_number", Value: req.StudentNumber},
		}

		// Update the student's document in the "students" collection.
		_, err := client.Collection("students").Doc(req.FirebaseID).Update(ctx, updates)
		if err != nil {
			http.Error(w, "Failed to update personal details: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Personal details updated successfully"))
	}
}
