// delete_goal.go
package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// DeleteGoalRequest defines the expected payload for deleting a goal.
type DeleteGoalRequest struct {
	FirebaseID string `json:"firebase_id"`
	College    string `json:"college"`
}

// DeleteGoalHandler returns an HTTP handler function that deletes a goal.
// It finds the student document in the "students" collection by the given firebase_id,
// then deletes the document (with document ID equal to the college name) from the subcollection "Goals".
func DeleteGoalHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req DeleteGoalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Validate required fields.
		if req.FirebaseID == "" || req.College == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Delete the goal document from the "Goals" subcollection.
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Goals").Doc(req.College).Delete(ctx)
		if err != nil {
			http.Error(w, "Failed to delete goal: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Goal deleted successfully"))
	}
}
