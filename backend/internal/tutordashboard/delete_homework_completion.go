// delete_event.go
package tutordashboard

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// DeleteEventRequest defines the payload required to delete an event.
type DeleteEventRequest struct {
	FirebaseID string `json:"firebase_id"` // The student's Firebase ID.
	EventID    string `json:"event_id"`    // The document ID (a date with dashes) for the event.
}

// DeleteEventHandler returns an HTTP handler function that deletes a given event.
func DeleteHomeworkCompletionHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS request.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var req DeleteEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Basic validation.
		if req.FirebaseID == "" || req.EventID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		// Delete the event document from the "Events" subcollection.
		_, err := client.Collection("students").Doc(req.FirebaseID).
			Collection("Homework Completion").Doc(req.EventID).Delete(ctx)
		if err != nil {
			http.Error(w, "Failed to delete event: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Event deleted successfully"))
	}
}
