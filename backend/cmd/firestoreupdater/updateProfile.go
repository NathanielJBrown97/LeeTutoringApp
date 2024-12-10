package firestoreupdater

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfileData struct {
	FirebaseID     string `json:"firebase_id"`    // The document ID and also stored in business.firebase_id
	StudentName    string `json:"studentName"`    // B4 name, may still be used in UI but not for the query
	StudentEmail   string `json:"studentEmail"`   // B2
	ParentEmail    string `json:"parentEmail"`    // B3
	Name           string `json:"name"`           // B4 full name of the student
	HighSchool     string `json:"highSchool"`     // B5
	Grade          string `json:"grade"`          // B6
	TestFocus      string `json:"testFocus"`      // B7
	Accommodations string `json:"accommodations"` // B8
	RegisteredTest string `json:"registeredTest"` // B9 (date string or "TBD")
	WhoSchedules   string `json:"whoSchedules"`   // B10
	Interests      string `json:"interests"`      // B11
	Availability   string `json:"availability"`   // B12 (not currently used)
	Notes          string `json:"notes"`          // B13
}

func (app *App) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	var profileData ProfileData
	err := json.NewDecoder(r.Body).Decode(&profileData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client := app.FirestoreClient

	// Use the FirebaseID as the document ID
	studentDocRef := client.Collection("students").Doc(profileData.FirebaseID)

	// Check if the document exists
	docSnap, err := studentDocRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get student doc: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !docSnap.Exists() {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	// Handle test appointment logic
	registeredForTest := false
	testDate := ""
	if profileData.RegisteredTest != "" && profileData.RegisteredTest != "TBD" {
		// If we have a real date
		registeredForTest = true
		testDate = profileData.RegisteredTest
	} else if profileData.RegisteredTest == "TBD" {
		// If TBD, registered_for_test should be false and test_date empty
		registeredForTest = false
		testDate = ""
	}
	// If empty, also false and no date (already defaulted)

	// Prepare the list of updates
	updates := []firestore.Update{
		// personal fields
		{Path: "personal.student_email", Value: profileData.StudentEmail},
		{Path: "personal.parent_email", Value: profileData.ParentEmail},
		{Path: "personal.name", Value: profileData.Name},
		{Path: "personal.high_school", Value: profileData.HighSchool},
		{Path: "personal.grade", Value: profileData.Grade},
		{Path: "personal.accommodations", Value: profileData.Accommodations},
		{Path: "personal.interests", Value: profileData.Interests},
		// availability not stored yet

		// business fields
		{Path: "business.test_focus", Value: profileData.TestFocus},
		{Path: "business.scheduler", Value: profileData.WhoSchedules},
		{Path: "business.notes", Value: profileData.Notes},
		{Path: "business.test_appointment.registered_for_test", Value: registeredForTest},
		{Path: "business.test_appointment.test_date", Value: testDate},
	}

	_, err = studentDocRef.Update(ctx, updates)
	if err != nil {
		log.Printf("Failed to update profile data in Firestore: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}
