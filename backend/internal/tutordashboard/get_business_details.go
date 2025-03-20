// get_business_details.go
package tutordashboard

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
)

// GetBusinessDetailsResponse defines the structure of the response payload.
type GetBusinessDetailsResponse struct {
	Business map[string]interface{} `json:"business"`
}

// GetBusinessDetailsHandler returns an HTTP handler that retrieves a student's business details.
func GetBusinessDetailsHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Expect firebase_id to be provided as a query parameter.
		firebaseID := r.URL.Query().Get("firebase_id")
		if firebaseID == "" {
			http.Error(w, "Missing firebase_id parameter", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		docRef := client.Collection("students").Doc(firebaseID)
		docSnap, err := docRef.Get(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving student document: %v", err), http.StatusInternalServerError)
			return
		}

		// Extract the "business" subdocument.
		var businessData map[string]interface{}
		if val, ok := docSnap.Data()["business"]; ok {
			if castVal, ok := val.(map[string]interface{}); ok {
				businessData = castVal
			} else {
				businessData = make(map[string]interface{})
			}
		} else {
			businessData = make(map[string]interface{})
		}

		response := GetBusinessDetailsResponse{
			Business: businessData,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
