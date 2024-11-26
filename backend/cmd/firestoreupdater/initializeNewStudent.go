package firestoreupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
)

type StudentData struct {
	Name              string `json:"name"`
	StudentEmail      string `json:"student_email"`
	StudentNumber     string `json:"student_number"`
	ParentEmail       string `json:"parent_email"`
	ParentNumber      string `json:"parent_number"`
	School            string `json:"school"`
	Grade             string `json:"grade"`
	Scheduler         string `json:"scheduler"`
	TestFocus         string `json:"test_focus"`
	Accommodations    string `json:"accommodations"`
	Interests         string `json:"interests"`
	Availability      string `json:"availability"`
	RegisteredForTest bool   `json:"registered_for_test"`
	TestDate          string `json:"test_date"`
}

func InitializeNewStudent(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON request body into studentData
	var studentData StudentData
	err := json.NewDecoder(r.Body).Decode(&studentData)
	if err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Initialize Firestore client
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("FIREBASE_PROJECT_ID"))
	if err != nil {
		http.Error(w, "Failed to create Firestore client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Generate a unique ID for the new student document
	docRef := client.Collection("students").NewDoc()
	studentID := docRef.ID

	// Prepare 'personal' data
	personalData := map[string]interface{}{
		"name":           studentData.Name,
		"student_email":  studentData.StudentEmail,
		"student_number": studentData.StudentNumber,
		"parent_email":   studentData.ParentEmail,
		"parent_number":  studentData.ParentNumber,
		"high_school":    studentData.School,
		"grade":          studentData.Grade,
		"accommodations": studentData.Accommodations,
		"interests":      studentData.Interests,
	}

	// Prepare 'test_appointment' data
	testAppointmentData := map[string]interface{}{
		"registered_for_test": studentData.RegisteredForTest,
		"test_date":           studentData.TestDate,
	}

	// Prepare 'business' data
	businessData := map[string]interface{}{
		"firebase_id":       studentID,
		"scheduler":         studentData.Scheduler,
		"test_focus":        studentData.TestFocus,
		"test_appointment":  testAppointmentData,
		"associated_tutors": []string{}, // Initialize as empty array
		"team_lead":         "",         // Initialize as empty string
		"remaining_hours":   0,          // Initialize as zero
		"lifetime_hours":    0,          // Initialize as zero
	}

	// Combine 'personal' and 'business' into the student document
	studentDocData := map[string]interface{}{
		"personal": personalData,
		"business": businessData,
	}

	// Write the student document to Firestore
	_, err = docRef.Set(ctx, studentDocData)
	if err != nil {
		http.Error(w, "Failed to save student data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success
	fmt.Fprintf(w, "Student %s initialized successfully with ID %s", studentData.Name, studentID)
}
