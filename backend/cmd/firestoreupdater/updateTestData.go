package firestoreupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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

	// Create document ID with quality prefix
	docID := fmt.Sprintf("%s %s %s", testData.Quality, testData.Test, testData.Date)

	// Initialize data to store
	data := map[string]interface{}{
		"date":       testData.Date,
		"test":       testData.Test,
		"type":       testData.Quality,
		"baseline":   testData.Baseline,
		"updated_at": time.Now().Format(time.RFC3339),
	}

	// Prepare scores subdocuments
	actScores := map[string]interface{}{
		"English":   nil,
		"Math":      nil,
		"Reading":   nil,
		"Science":   nil,
		"ACT_Total": nil,
	}
	satScores := map[string]interface{}{
		"EBRW":      nil,
		"Math":      nil,
		"Reading":   nil,
		"Writing":   nil,
		"SAT_Total": nil,
	}

	scores := testData.Scores

	if testData.Test == "ACT" {
		// Populate ACT_Scores
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

		if actTotal, ok := scores["actTotal"].(float64); ok {
			actScores["ACT_Total"] = actTotal
		}

	} else if testData.Test == "SAT" || testData.Test == "PSAT" {
		// Populate SAT_Scores
		if ebrw, ok := scores["ebrw"]; ok {
			satScores["EBRW"] = ebrw
		}
		if math, ok := scores["math"]; ok {
			satScores["Math"] = math
		}
		if reading, ok := scores["reading"]; ok {
			satScores["Reading"] = reading
		}
		if writing, ok := scores["writing"]; ok {
			satScores["Writing"] = writing
		}

		if satTotal, ok := scores["satTotal"].(float64); ok {
			satScores["SAT_Total"] = satTotal
		}
	}

	// Add scores subdocuments to data
	data["ACT_Scores"] = actScores
	data["SAT_Scores"] = satScores

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
