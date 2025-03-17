// create_goal.go
package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// CreateGoalRequest defines the expected payload for creating a new goal.
type CreateGoalRequest struct {
	FirebaseID     string            `json:"firebase_id"`
	College        string            `json:"college"`
	ActPercentiles map[string]string `json:"act_percentiles"`
	SatPercentiles map[string]string `json:"sat_percentiles"`
}

// CreateGoalHandler returns an HTTP handler function that processes a new goal creation request.
// It finds the student's document in the "students" collection using FirebaseID, then creates a document in
// the subcollection "Goals" (with the document ID equal to the college name) containing the fields:
// "College" (string), "SAT_percentiles" (array), and "ACT_percentiles" (array).
func CreateGoalHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req CreateGoalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" || req.College == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Build the data for the new goal document.
		goalData := map[string]interface{}{
			"College":         req.College,
			"SAT_percentiles": []string{req.SatPercentiles["p25"], req.SatPercentiles["p50"], req.SatPercentiles["p75"]},
			"ACT_percentiles": []string{req.ActPercentiles["p25"], req.ActPercentiles["p50"], req.ActPercentiles["p75"]},
		}

		// Write the new goal document in the student's subcollection "Goals" (using the college name as the document ID).
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Goals").Doc(req.College).Set(ctx, goalData)
		if err != nil {
			http.Error(w, "Failed to create new goal: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Goal created successfully"))
	}
}
