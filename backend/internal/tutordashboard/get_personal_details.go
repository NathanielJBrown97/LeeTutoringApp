// get_personal_details.go
package tutordashboard

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
)

// GetPersonalDetailsResponse defines the structure of the response payload.
type GetPersonalDetailsResponse struct {
	Personal map[string]interface{} `json:"personal"`
}

// GetPersonalDetailsHandler returns an HTTP handler that retrieves a student's personal details.
func GetPersonalDetailsHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Expect firebase_id as a query parameter.
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

		// Extract the "personal" subdocument.
		var personalData map[string]interface{}
		if val, ok := docSnap.Data()["personal"]; ok {
			if castVal, ok := val.(map[string]interface{}); ok {
				personalData = castVal
			} else {
				personalData = make(map[string]interface{})
			}
		} else {
			personalData = make(map[string]interface{})
		}

		response := GetPersonalDetailsResponse{
			Personal: personalData,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
