package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// DeleteTestDataRequest defines the expected payload for deleting a test data entry.
type DeleteTestDataRequest struct {
	FirebaseID string `json:"firebase_id"`  // The student's Firebase ID.
	TestDataID string `json:"test_data_id"` // The ID of the test data document (e.g., "Official ACT 10-26-2024")
}

// DeleteTestDataHandler returns an HTTP handler function that processes a delete test data request.
func DeleteTestDataHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req DeleteTestDataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" || req.TestDataID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Delete the test data document from the "Test Data" subcollection.
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Test Data").Doc(req.TestDataID).Delete(ctx)
		if err != nil {
			http.Error(w, "Failed to delete test data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test data deleted successfully"))
	}
}
