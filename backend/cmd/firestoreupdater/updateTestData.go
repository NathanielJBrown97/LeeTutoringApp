package firestoreupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/api/iterator"
)

type TestData struct {
	StudentName string                 `json:"studentName"`
	Date        string                 `json:"date"`    // Date as provided, e.g., "10-14-2024"
	Test        string                 `json:"test"`    // "ACT", "SAT", or "PSAT"
	Quality     string                 `json:"quality"` // Will be stored as "type" in the database
	Baseline    bool                   `json:"baseline"`
	Scores      map[string]interface{} `json:"scores"`
}

func (app *App) UpdateTestDataHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	var testData TestData
	err := json.NewDecoder(r.Body).Decode(&testData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client := app.FirestoreClient

	// Query the 'students' collection for the student with matching name
	iter := client.Collection("students").Where("personal.name", "==", testData.StudentName).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query Firestore: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	studentDocRef := doc.Ref

	// Access the 'Test Data' subcollection
	testDataCollection := studentDocRef.Collection("Test Data")

	// Use the date as provided without any formatting
	docID := fmt.Sprintf("%s %s", testData.Test, testData.Date)

	// Prepare data to be stored
	data := map[string]interface{}{
		"date":     testData.Date,
		"test":     testData.Test,
		"type":     testData.Quality, // Store 'quality' as 'type'
		"baseline": testData.Baseline,
	}

	// Process the 'scores' map to fit into the 'ACT' or 'SAT' field
	scores := testData.Scores

	// Initialize the scores data
	var actScores map[string]interface{}
	var satScores map[string]interface{}

	if testData.Test == "ACT" {
		// We have an ACT test
		actScores = make(map[string]interface{})

		// Extract ACT scores
		if total, ok := scores["composite"]; ok {
			actScores["ACT Total"] = total
		}
		if english, ok := scores["english"]; ok {
			actScores["English"] = english
		}
		if math, ok := scores["math"]; ok {
			actScores["Math"] = math
		}
		if reading, ok := scores["reading"]; ok {
			actScores["Reading"] = reading
		}
		if science, ok := scores["science"]; ok {
			actScores["Science"] = science
		}

		data["ACT"] = actScores

		// Check for alternate test's total score (SAT Total)
		if satTotal, ok := scores["sat_total"]; ok {
			satScores = make(map[string]interface{})
			satScores["SAT Total"] = satTotal
			data["SAT"] = satScores
		}

	} else if testData.Test == "SAT" || testData.Test == "PSAT" {
		// We have an SAT or PSAT test
		satScores = make(map[string]interface{})

		// Extract SAT scores
		if total, ok := scores["total"]; ok {
			satScores["SAT Total"] = total
		}
		if ebrw, ok := scores["ebrw"]; ok {
			satScores["EBRW"] = ebrw
		}
		if math, ok := scores["math"]; ok {
			satScores["Math"] = math
		}

		data["SAT"] = satScores

		// Check for alternate test's total score (ACT Total)
		if actTotal, ok := scores["act_total"]; ok {
			actScores = make(map[string]interface{})
			actScores["ACT Total"] = actTotal
			data["ACT"] = actScores
		}
	} else {
		// Handle unexpected test types
		log.Printf("Unknown test type: %s", testData.Test)
		http.Error(w, "Unknown test type", http.StatusBadRequest)
		return
	}

	// Create or update the document
	_, err = testDataCollection.Doc(docID).Set(ctx, data)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}
