// edit_homework_completion.go
package tutordashboard

import (
	"encoding/json"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
)

// EditHomeworkCompletionRequest defines the expected payload for editing a homework completion entry.
type EditHomeworkCompletionRequest struct {
	FirebaseID         string `json:"firebase_id"`         // The student's Firebase ID.
	Attendance         string `json:"attendance"`          // e.g., "On Time", "Late", "Ended Early", "No Show".
	Date               string `json:"date"`                // Date of homework completion (e.g., "02/26/2025").
	Duration           string `json:"duration"`            // Duration value as a string (e.g., "0.25", "0.50", "0.75", etc.).
	Feedback           string `json:"feedback"`            // Feedback text.
	PercentageComplete string `json:"percentage_complete"` // Percentage complete value as a string.
	Engagement         string `json:"engagement"`          // Engagement level as a string.
	Tutor              string `json:"tutor"`               // Tutor's first name.
	Timestamp          string `json:"timestamp"`           // Timestamp in ISO format (e.g., "2025-02-11T00:40:05Z").
}

// EditHomeworkCompletionHandler returns an HTTP handler function that processes an edit homework completion request.
func EditHomeworkCompletionHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req EditHomeworkCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" || req.Date == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Build the document ID by replacing slashes with dashes in the date.
		docID := strings.ReplaceAll(req.Date, "/", "-")

		// Build the update data.
		updates := []firestore.Update{
			{Path: "attendance", Value: req.Attendance},
			{Path: "date", Value: req.Date},
			{Path: "duration", Value: req.Duration},
			{Path: "feedback", Value: req.Feedback},
			{Path: "percentage_complete", Value: req.PercentageComplete},
			{Path: "engagement", Value: req.Engagement},
			{Path: "tutor", Value: req.Tutor},
			{Path: "timestamp", Value: req.Timestamp},
		}

		// Update the homework completion document in the "Homework Completion" subcollection.
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Homework Completion").Doc(docID).Update(ctx, updates)
		if err != nil {
			http.Error(w, "Failed to update homework completion: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Homework completion updated successfully"))
	}
}
