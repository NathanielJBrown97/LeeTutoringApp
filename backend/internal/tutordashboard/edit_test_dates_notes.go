// edit_test_dates_notes.go
package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// EditTestDatesNotesRequest defines the expected payload for editing test data notes.
type EditTestDatesNotesRequest struct {
	FirebaseID   string `json:"firebase_id"`   // The student's Firebase ID.
	DocumentName string `json:"document_name"` // The document name in the TestData subcollection (e.g., "ACT 7-12-2025").
	Notes        string `json:"notes"`         // The new notes value to update.
}

// EditTestDatesNotesHandler returns an HTTP handler function that processes a request to update notes in the TestData subcollection.
func EditTestDatesNotesHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()
		var req EditTestDatesNotesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" || req.DocumentName == "" || req.Notes == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Prepare the update for the 'notes' field.
		updates := []firestore.Update{
			{Path: "notes", Value: req.Notes},
		}

		// Update the document in the "TestData" subcollection.
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Test Dates").Doc(req.DocumentName).Update(ctx, updates)
		if err != nil {
			http.Error(w, "Failed to update test data notes: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test data notes updated successfully"))
	}
}
