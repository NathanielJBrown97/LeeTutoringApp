// backend/internal/parent/student_intake.go

package parent

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/NathanielJBrown97/LeeTutoringApp/internal/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentInfo struct {
	StudentID   string `json:"studentId"`
	StudentName string `json:"studentName"`
	CanLink     bool   `json:"canLink"`
}

// StudentIntakeHandler handles the submission of student IDs
func (a *App) StudentIntakeHandler(w http.ResponseWriter, r *http.Request) {
	// Authentication is handled via middleware
	userID, err := middleware.ExtractUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}
	// Optionally, use userID for logging or other purposes
	log.Printf("User ID in StudentIntakeHandler: %s", userID)

	// Parse the JSON request body
	var requestData struct {
		StudentIDs []string `json:"studentIds"`
	}
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Failed to parse JSON request body", http.StatusBadRequest)
		return
	}

	if len(requestData.StudentIDs) == 0 {
		http.Error(w, "No student IDs provided", http.StatusBadRequest)
		return
	}

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Prepare to store student names or errors
	var studentInfos []StudentInfo

	// Loop through the student IDs and fetch names from Firestore
	for _, studentID := range requestData.StudentIDs {
		if studentID == "" {
			continue
		}

		// Access the student document in Firestore
		docRef := firestoreClient.Collection("students").Doc(studentID)
		doc, err := docRef.Get(context.Background())
		if err != nil {
			if status.Code(err) == codes.NotFound {
				log.Printf("Student ID %s: Document not found", studentID)
				studentInfos = append(studentInfos, StudentInfo{
					StudentID:   studentID,
					StudentName: "Not found",
					CanLink:     false, // Flag that this student cannot be linked
				})
				continue
			} else {
				log.Printf("Error retrieving student ID %s: %v", studentID, err)
				http.Error(w, "Failed to retrieve student data", http.StatusInternalServerError)
				return
			}
		}

		// Access the personal sub-document data as a map
		personalData, ok := doc.Data()["personal"].(map[string]interface{})
		if !ok {
			log.Printf("Student ID %s: Personal data not found or is not a map", studentID)
			studentInfos = append(studentInfos, StudentInfo{
				StudentID:   studentID,
				StudentName: "No personal data",
				CanLink:     false, // Flag that this student cannot be linked
			})
			continue
		}

		// Extract the "name" field from the personal data
		studentName, ok := personalData["name"].(string)
		if !ok || studentName == "" {
			log.Printf("Student ID %s: Name field not found or is empty", studentID)
			studentInfos = append(studentInfos, StudentInfo{
				StudentID:   studentID,
				StudentName: "No name field",
				CanLink:     false, // Flag that this student cannot be linked
			})
		} else {
			studentInfos = append(studentInfos, StudentInfo{
				StudentID:   studentID,
				StudentName: studentName,
				CanLink:     true, // Flag that this student can be linked
			})
		}
	}

	// Return the student information as JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(struct {
		StudentInfos []StudentInfo `json:"studentInfos"`
	}{
		StudentInfos: studentInfos,
	})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
