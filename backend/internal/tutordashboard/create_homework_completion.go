// create_homework_completion.go
package tutordashboard

import (
	"encoding/json"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
)

// CreateHomeworkCompletionRequest defines the expected payload for creating a new homework completion entry.
type CreateHomeworkCompletionRequest struct {
	FirebaseID         string `json:"firebase_id"`         // The student's Firebase ID.
	Attendance         string `json:"attendance"`          // e.g., "On Time", "Late", "Ended Early", "No Show".
	Date               string `json:"date"`                // Date of homework completion (stored with slashes, e.g., "02/26/2025").
	Duration           string `json:"duration"`            // Duration value as a string (e.g., "0.25", "0.50", "0.75", etc.).
	Feedback           string `json:"feedback"`            // Feedback text.
	PercentageComplete string `json:"percentage_complete"` // Percentage complete value as a string.
	Engagement         string `json:"engagement"`          // Engagement level as a string.
	Tutor              string `json:"tutor"`               // Tutor's first name.
	Timestamp          string `json:"timestamp"`           // Timestamp in ISO format (e.g., "2025-02-11T00:40:05Z").
}

// CreateHomeworkCompletionHandler returns an HTTP handler function that processes a new homework completion creation request.
func CreateHomeworkCompletionHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req CreateHomeworkCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" || req.Date == "" || req.Attendance == "" || req.Duration == "" || req.Timestamp == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Build the document ID by replacing slashes with dashes in the date.
		docID := strings.ReplaceAll(req.Date, "/", "-")

		// Build the homework completion object.
		homeworkData := map[string]interface{}{
			"attendance":          req.Attendance,
			"date":                req.Date, // Stored with slashes.
			"duration":            req.Duration,
			"feedback":            req.Feedback,
			"percentage_complete": req.PercentageComplete,
			"engagement":          req.Engagement,
			"tutor":               req.Tutor,
			"timestamp":           req.Timestamp,
		}

		// Write the new homework completion document in the "Homework Completion" subcollection.
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Homework Completion").Doc(docID).Set(ctx, homeworkData)
		if err != nil {
			http.Error(w, "Failed to create homework completion: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Homework completion created successfully"))
	}
}
