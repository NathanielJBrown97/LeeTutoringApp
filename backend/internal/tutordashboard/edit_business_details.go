// edit_business_details.go
package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// EditBusinessDetailsRequest defines the expected payload for editing a student's business details.
type EditBusinessDetailsRequest struct {
	FirebaseID       string   `json:"firebase_id"`       // The student's Firebase ID.
	AssociatedTutors []string `json:"associated_tutors"` // Array of tutor names (e.g., ["Ben", "Kyra"]).
	Scheduler        string   `json:"scheduler"`         // e.g., "Either Parent", "Mother", "Father", "Student".
	Status           string   `json:"status"`            // e.g., "Awaiting Results", "Active", "Inactive".
	TeamLead         string   `json:"team_lead"`         // e.g., "Edward", "Eli", "Ben", "Kieran", "Kyra", "Patrick".
	TestFocus        string   `json:"test_focus"`        // e.g., "ACT" or "SAT".
	// TestAppointment holds details about test registration.
	TestAppointment struct {
		RegisteredForTest bool   `json:"registered_for_test"` // true if registered for test.
		TestDate          string `json:"test_date"`           // Test date in "YYYY-MM-DD" format.
	} `json:"test_appointment"`
	Notes string `json:"notes"` // Additional notes.
}

// EditBusinessDetailsHandler returns an HTTP handler function that processes an edit business details request.
func EditBusinessDetailsHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req EditBusinessDetailsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" {
			http.Error(w, "Missing required field: firebase_id", http.StatusBadRequest)
			return
		}

		// Build the update data. Note: Using field paths to update the nested "business" subdocument.
		updates := []firestore.Update{
			{Path: "business.associated_tutors", Value: req.AssociatedTutors},
			{Path: "business.scheduler", Value: req.Scheduler},
			{Path: "business.status", Value: req.Status},
			{Path: "business.team_lead", Value: req.TeamLead},
			{Path: "business.test_focus", Value: req.TestFocus},
			{Path: "business.test_appointment", Value: req.TestAppointment},
			{Path: "business.notes", Value: req.Notes},
		}

		// Update the student's document in the "students" collection.
		_, err := client.Collection("students").Doc(req.FirebaseID).Update(ctx, updates)
		if err != nil {
			http.Error(w, "Failed to update business details: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Business details updated successfully"))
	}
}
