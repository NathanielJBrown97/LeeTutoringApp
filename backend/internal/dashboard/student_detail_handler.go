package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// StudentDetailResponse represents the detailed data for a student
type StudentDetailResponse struct {
	ID                 string                   `json:"id"`
	Personal           map[string]interface{}   `json:"personal"`
	Business           map[string]interface{}   `json:"business"`
	HomeworkCompletion []map[string]interface{} `json:"homeworkCompletion"`
	TestData           []map[string]interface{} `json:"testData"`
	TestDates          []map[string]interface{} `json:"testDates"`
	Goals              []map[string]interface{} `json:"goals"`
}

// StudentDetailHandler handles the GET /api/students/{student_id} endpoint
func (a *App) StudentDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user ID from the JWT token
	userID, _ := a.getParentCredentials(r)
	if userID == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
		return
	}

	// Fetch associated_students from Firestore
	associatedStudents, err := a.getAssociatedStudents(userID)
	if err != nil {
		log.Printf("Error fetching associated students: %v", err)
		http.Error(w, "Unable to fetch associated students", http.StatusInternalServerError)
		return
	}
	if len(associatedStudents) == 0 {
		http.Error(w, "No associated students found", http.StatusUnauthorized)
		return
	}

	// Get the student_id from the URL
	vars := mux.Vars(r)
	studentID := vars["student_id"]
	if studentID == "" {
		http.Error(w, "Student ID is required", http.StatusBadRequest)
		return
	}

	// Verify that the student is associated with the parent
	isAssociated := false
	for _, sID := range associatedStudents {
		if sID == studentID {
			isAssociated = true
			break
		}
	}
	if !isAssociated {
		http.Error(w, "Unauthorized access to student data", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Fetch student data
	studentDoc, err := a.FirestoreClient.Collection("students").Doc(studentID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching student document with ID %s: %v", studentID, err)
		http.Error(w, "Error fetching student data", http.StatusInternalServerError)
		return
	}

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
	homeworkCompletionDocs, err := studentDoc.Ref.Collection("Homework Completion").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching 'Homework Completion' subcollection: %v", err)
		studentData.HomeworkCompletion = []map[string]interface{}{}
	} else {
		var homeworkCompletion []map[string]interface{}
		for _, doc := range homeworkCompletionDocs {
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
	testDataDocs, err := studentDoc.Ref.Collection("Test Data").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching 'Test Data' subcollection: %v", err)
		studentData.TestData = []map[string]interface{}{}
	} else {
		var testData []map[string]interface{}
		for _, doc := range testDataDocs {
			data := doc.Data()
			data["id"] = doc.Ref.ID

			// Convert ACT_Scores map into an array if present
			if actScoresMap, ok := data["ACT_Scores"].(map[string]interface{}); ok {
				// The frontend expects them in a specific order:
				// For ACT_Scores, we have 'ACT_Total', 'English', 'Math', 'Reading', 'Science'
				// The frontend code is indexing them in a known order. Let's assume we provide them in:
				// [English, Math, Reading, Science, ACT_Total]
				actArray := []interface{}{
					actScoresMap["English"],
					actScoresMap["Math"],
					actScoresMap["Reading"],
					actScoresMap["Science"],
					actScoresMap["ACT_Total"],
				}
				data["ACT_Scores"] = actArray
			}

			// Convert SAT_Scores map into an array if present
			if satScoresMap, ok := data["SAT_Scores"].(map[string]interface{}); ok {
				// For SAT_Scores, we have 'EBRW', 'Math', 'Reading', 'SAT_Total', 'Writing'
				// The frontend expects them as: [EBRW, Math, Reading, Writing, SAT_Total]
				satArray := []interface{}{
					satScoresMap["EBRW"],
					satScoresMap["Math"],
					satScoresMap["Reading"],
					satScoresMap["Writing"],
					satScoresMap["SAT_Total"],
				}
				data["SAT_Scores"] = satArray
			}

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

	// Return the student data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(studentData)
}
