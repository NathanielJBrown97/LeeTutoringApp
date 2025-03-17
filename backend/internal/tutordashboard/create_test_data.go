// create_test_data.go
package tutordashboard

import (
	"encoding/json"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
)

// CreateTestDataRequest defines the expected payload for creating a new test data entry.
type CreateTestDataRequest struct {
	FirebaseID string             `json:"firebase_id"` // The student's Firebase ID.
	Date       string             `json:"date"`        // Date of test as string (e.g., "10/26/2024")
	Baseline   bool               `json:"baseline"`    // Baseline value.
	Test       string             `json:"test"`        // Test name (e.g., ACT, SAT, PSAT, PACT)
	Type       string             `json:"type"`        // Type (e.g., Official, Unofficial, SS Official, SS Unofficial)
	ACT_Scores map[string]float64 `json:"ACT_Scores"`  // ACT scores: ACT_Total, English, Math, Reading, Science.
	SAT_Scores map[string]float64 `json:"SAT_Scores"`  // SAT scores: SAT_Total, EBRW, Math, Reading, Writing.
}

// CreateTestDataHandler returns an HTTP handler function that processes a new test data creation request.
func CreateTestDataHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		ctx := r.Context()

		var req CreateTestDataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" || req.Date == "" || req.Test == "" || req.Type == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Build the document ID.
		// Convert the date string from "MM/DD/YYYY" to "MM-DD-YYYY" for the document ID.
		docID := req.Type + " " + req.Test + " " + strings.ReplaceAll(req.Date, "/", "-")

		// Build the test data object.
		testData := map[string]interface{}{
			"date":       req.Date,      // Stored with slashes.
			"baseline":   req.Baseline,
			"test":       req.Test,
			"type":       req.Type,
			"ACT_Scores": req.ACT_Scores,
			"SAT_Scores": req.SAT_Scores,
		}

		// Write the new test data document in the "Test Data" subcollection.
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Test Data").Doc(docID).Set(ctx, testData)
		if err != nil {
			http.Error(w, "Failed to create test data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test data created successfully"))
	}
}
