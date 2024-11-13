// backend/internal/parent/student_link_confirmation.go

package parent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/middleware"
)

func sliceStringsToInterfaces(slice []string) []interface{} {
	interfaces := make([]interface{}, len(slice))
	for i, v := range slice {
		interfaces[i] = v
	}
	return interfaces
}

// ConfirmLinkStudentsHandler handles the confirmation of linking students
func (a *App) ConfirmLinkStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// Authentication is handled via middleware
	userID, err := middleware.ExtractUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}

	// Parse the JSON request body
	var requestData struct {
		ConfirmedStudentIDs []string `json:"confirmedStudentIds"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Failed to parse JSON request body", http.StatusBadRequest)
		return
	}

	if len(requestData.ConfirmedStudentIDs) == 0 {
		http.Error(w, "No confirmed student IDs provided", http.StatusBadRequest)
		return
	}

	// Convert []string to []interface{}
	confirmedStudentIDsInterface := sliceStringsToInterfaces(requestData.ConfirmedStudentIDs)

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Reference to the parent's document using the userID from the session
	parentDocRef := firestoreClient.Collection("parents").Doc(userID)

	// Update the parent's associated_students field
	_, err = parentDocRef.Set(context.Background(), map[string]interface{}{
		"associated_students": firestore.ArrayUnion(confirmedStudentIDsInterface...),
	}, firestore.MergeAll)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to link students: %v", err), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Students linked successfully"}`))
}
