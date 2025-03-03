// backend/internal/tutordashboard/tutor_student_detail_handler.go
package tutordashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
)

// StudentDetailResponse represents the detailed data for a student.
type StudentDetailResponse struct {
	ID                 string                   `json:"id"`
	Personal           map[string]interface{}   `json:"personal"`
	Business           map[string]interface{}   `json:"business"`
	HomeworkCompletion []map[string]interface{} `json:"homeworkCompletion"`
	TestData           []map[string]interface{} `json:"testData"`
	TestDates          []map[string]interface{} `json:"testDates"`
	Goals              []map[string]interface{} `json:"goals"`
}

// App represents your application context. It should include FirestoreClient and any credential helper functions.
type App struct {
	FirestoreClient *firestore.Client
	// Other fields such as logger, config, etc.
}

// getTutorCredentials extracts the tutor's user ID and email from the request context (for example, from the JWT token).
// For simplicity, this is a stub. In production, you should implement proper JWT verification.
func (a *App) getTutorCredentials(r *http.Request) (string, string, error) {
	// TODO: Extract tutor credentials from request context / JWT token.
	// For example purposes, let's assume query parameters "tutorUserID" and "tutorEmail" are passed.
	tutorUserID := r.URL.Query().Get("tutorUserID")
	tutorEmail := r.URL.Query().Get("tutorEmail")
	if tutorUserID == "" || tutorEmail == "" {
		return "", "", http.ErrNoCookie
	}
	return tutorUserID, tutorEmail, nil
}

// TutorStudentDetailHandler handles GET /api/tutor/students/{student_id} requests.
// It only returns the student details if the student is in the tutor's "Associated Students" subcollection.
func (a *App) TutorStudentDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract tutor credentials.
	tutorUserID, _, err := a.getTutorCredentials(r)
	if err != nil || tutorUserID == "" {
		http.Error(w, "Unable to identify tutor user", http.StatusUnauthorized)
		return
	}

	// Get student ID from URL path variables.
	vars := mux.Vars(r)
	studentID := vars["student_id"]
	if studentID == "" {
		http.Error(w, "Student ID is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Check if the student exists in the tutor's "Associated Students" subcollection.
	_, err = a.FirestoreClient.Collection("tutors").Doc(tutorUserID).
		Collection("Associated Students").Doc(studentID).Get(ctx)
	if err != nil {
		log.Printf("Student %s not associated with tutor %s: %v", studentID, tutorUserID, err)
		http.Error(w, "Unauthorized access to student data", http.StatusUnauthorized)
		return
	}
	// (Optional) You can log or verify data from associatedStudentDoc.Data() if needed.
	log.Printf("Tutor %s is associated with student %s", tutorUserID, studentID)

	// Fetch student details from the "students" collection.
	studentDoc, err := a.FirestoreClient.Collection("students").Doc(studentID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching student document with ID %s: %v", studentID, err)
		http.Error(w, "Error fetching student data", http.StatusInternalServerError)
		return
	}

	// Build the student detail response.
	studentData := StudentDetailResponse{
		ID: studentID,
	}

	// Personal Data
	if personalData, ok := studentDoc.Data()["personal"].(map[string]interface{}); ok {
		personalResponse := make(map[string]interface{})
		if name, ok := personalData["name"]; ok {
			personalResponse["name"] = name
		}
		if accommodations, ok := personalData["accommodations"]; ok {
			personalResponse["accommodations"] = accommodations
		}
		if grade, ok := personalData["grade"]; ok {
			personalResponse["grade"] = grade
		}
		if highSchool, ok := personalData["high_school"]; ok {
			personalResponse["high_school"] = highSchool
		}
		if parentEmail, ok := personalData["parent_email"]; ok {
			personalResponse["parent_email"] = parentEmail
		}
		if studentEmail, ok := personalData["student_email"]; ok {
			personalResponse["student_email"] = studentEmail
		}
		studentData.Personal = personalResponse
	} else {
		studentData.Personal = make(map[string]interface{})
	}

	// Business Data
	if businessData, ok := studentDoc.Data()["business"].(map[string]interface{}); ok {
		businessResponse := make(map[string]interface{})
		if lifetimeHours, ok := businessData["lifetime_hours"]; ok {
			businessResponse["lifetime_hours"] = lifetimeHours
		}
		if registeredTests, ok := businessData["registered_tests"]; ok {
			businessResponse["registered_tests"] = registeredTests
		}
		if remainingHours, ok := businessData["remaining_hours"]; ok {
			businessResponse["remaining_hours"] = remainingHours
		}
		if status, ok := businessData["status"]; ok {
			businessResponse["status"] = status
		}
		if teamLead, ok := businessData["team_lead"]; ok {
			businessResponse["team_lead"] = teamLead
		}
		if testFocus, ok := businessData["test_focus"]; ok {
			businessResponse["test_focus"] = testFocus
		}
		if associatedTutors, ok := businessData["associated_tutors"]; ok {
			businessResponse["associated_tutors"] = associatedTutors
		}
		studentData.Business = businessResponse
	} else {
		studentData.Business = make(map[string]interface{})
	}

	// Homework Completion Subcollection
	homeworkDocs, err := studentDoc.Ref.Collection("Homework Completion").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching 'Homework Completion' subcollection: %v", err)
		studentData.HomeworkCompletion = []map[string]interface{}{}
	} else {
		var homeworkCompletion []map[string]interface{}
		for _, doc := range homeworkDocs {
			data := doc.Data()
			homeworkData := map[string]interface{}{
				"id":         doc.Ref.ID,
				"attendance": data["attendance"],
				"date":       data["date"],
				"duration":   data["duration"],
				"feedback":   data["feedback"],
				"percentage": data["percentage_complete"],
				"timestamp":  data["timestamp"],
				"tutor":      data["tutor"],
			}
			homeworkCompletion = append(homeworkCompletion, homeworkData)
		}
		studentData.HomeworkCompletion = homeworkCompletion
	}

	// Test Data Subcollection
	testDocs, err := studentDoc.Ref.Collection("Test Data").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching 'Test Data' subcollection: %v", err)
		studentData.TestData = []map[string]interface{}{}
	} else {
		var testData []map[string]interface{}
		for _, doc := range testDocs {
			data := doc.Data()
			data["id"] = doc.Ref.ID
			// Optional: Convert ACT_Scores and SAT_Scores maps to arrays if required by the frontend.
			testData = append(testData, data)
		}
		studentData.TestData = testData
	}

	// Test Dates Subcollection
	testDatesDocs, err := studentDoc.Ref.Collection("Test Dates").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching 'Test Dates' subcollection: %v", err)
		studentData.TestDates = []map[string]interface{}{}
	} else {
		var testDates []map[string]interface{}
		for _, doc := range testDatesDocs {
			data := doc.Data()
			data["id"] = doc.Ref.ID
			testDates = append(testDates, data)
		}
		studentData.TestDates = testDates
	}

	// Goals Subcollection
	goalsDocs, err := studentDoc.Ref.Collection("Goals").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching 'Goals' subcollection: %v", err)
		studentData.Goals = []map[string]interface{}{}
	} else {
		var goals []map[string]interface{}
		for _, doc := range goalsDocs {
			data := doc.Data()
			data["id"] = doc.Ref.ID
			goals = append(goals, data)
		}
		studentData.Goals = goals
	}

	// Return the student detail as JSON.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(studentData)
}
