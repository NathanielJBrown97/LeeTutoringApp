// backend/internal/dashboard/associated_students_handler.go

package dashboard

import (
	"encoding/json"
	"log"
	"net/http"
)

// AssociatedStudentsResponse represents the response structure
type AssociatedStudentsResponse struct {
	AssociatedStudents []string `json:"associatedStudents"`
}

// AssociatedStudentsHandler handles the GET /api/associated-students endpoint
func (a *App) AssociatedStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve credentials from the JWT token
	userID, _ := a.getParentCredentials(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch associated_students from Firestore
	associatedStudents, err := a.getAssociatedStudents(userID)
	if err != nil {
		log.Printf("Error fetching associated students: %v", err)
		http.Error(w, "Unable to fetch associated students", http.StatusInternalServerError)
		return
	}

	// Return the associated students
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AssociatedStudentsResponse{
		AssociatedStudents: associatedStudents,
	})
}
