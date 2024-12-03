package firestoreupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	ClassroomID       string `json:"classroom_id"`
	DriveURL          string `json:"drive_url"`
	// FirestoreID       string `json:"firestore_id"` // Commented out; will be set internally
}

type ResponseData struct {
	Message   string `json:"message"`
	StudentID string `json:"student_id"`
}

func InitializeNewStudent(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON request body into studentData
	var studentData StudentData
	err := json.NewDecoder(r.Body).Decode(&studentData)
	if err != nil {
		log.Printf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Log the received student data
	log.Printf("Received student data: %+v", studentData)

	// Initialize Firestore client
	ctx := context.Background()
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		log.Println("FIREBASE_PROJECT_ID environment variable is not set")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Firestore client: %v", err)
		http.Error(w, "Failed to create Firestore client: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Generate a unique ID for the new student document
	docRef := client.Collection("students").NewDoc()
	studentID := docRef.ID

	// Log the generated student ID
	log.Printf("Generated student ID: %s", studentID)

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
		// "firestore_id":   studentID, // Optionally include firestore_id
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
		"classroom_id":      studentData.ClassroomID,
		"drive_url":         studentData.DriveURL,
	}

	// Combine 'personal' and 'business' into the student document
	studentDocData := map[string]interface{}{
		"personal": personalData,
		"business": businessData,
	}

	// Log the data to be written to Firestore
	log.Printf("studentDocData to be written: %+v", studentDocData)

	// Write the student document to Firestore
	_, err = docRef.Set(ctx, studentDocData)
	if err != nil {
		log.Printf("Failed to save student data: %v", err)
		http.Error(w, "Failed to save student data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the JSON response
	responseData := ResponseData{
		Message:   fmt.Sprintf("Student %s initialized successfully", studentData.Name),
		StudentID: studentID,
	}

	// Set response header to application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send the JSON response
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
	}

	// Log successful completion
	log.Printf("Successfully initialized student: %s", studentData.Name)
}
