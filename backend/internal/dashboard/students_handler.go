// backend/internal/dashboard/students_handler.go

package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// StudentListResponse represents the response structure for the list of students
type StudentListResponse struct {
	Students []Student `json:"students"`
}

// StudentsHandler handles the GET /api/students endpoint
func (a *App) StudentsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the parent document ID from the session
	parentDocumentID, _ := a.getParentCredentials(r)
	if parentDocumentID == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Fetch the parent document
	parentDoc, err := a.FirestoreClient.Collection("parents").Doc(parentDocumentID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching parent document: %v", err)
		http.Error(w, "Error fetching parent data", http.StatusInternalServerError)
		return
	}

	// Extract associated_students array
	associatedStudentsInterface, ok := parentDoc.Data()["associated_students"].([]interface{})
	if !ok || len(associatedStudentsInterface) == 0 {
		// No associated students found
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StudentListResponse{Students: []Student{}})
		return
	}

	var students []Student

	for _, s := range associatedStudentsInterface {
		studentID, ok := s.(string)
		if !ok {
			log.Println("Error casting student ID to string")
			continue
		}

		// Access the student document in Firestore
		studentDoc, err := a.FirestoreClient.Collection("students").Doc(studentID).Get(ctx)
		if err != nil {
			log.Printf("Error fetching student document with ID %s: %v", studentID, err)
			continue
		}

		personalData, ok := studentDoc.Data()["personal"].(map[string]interface{})
		if !ok {
			log.Printf("Error fetching 'personal' data for student ID %s", studentID)
			continue
		}

		name, nameOk := personalData["name"].(string)
		if !nameOk {
			log.Printf("Name not found for student ID %s", studentID)
			continue
		}

		students = append(students, Student{ID: studentID, Name: name})
	}

	// Return the list of students
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StudentListResponse{Students: students})
}
