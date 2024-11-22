package firestoreupdater

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"google.golang.org/api/iterator"
	"google.golang.org/api/sheets/v4"
)

type GoalsPayload struct {
	StudentName string   `json:"studentName"`
	Goals       []string `json:"goals"`
}

func (app *App) UpdateGoalsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON payload
	var payload GoalsPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client := app.FirestoreClient

	// Find the student in 'students' collection where 'personal.name' == payload.StudentName
	iter := client.Collection("students").Where("personal.name", "==", payload.StudentName).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query Firestore: %v", err)
		http.Error(w, "Failed to query Firestore: "+err.Error(), http.StatusInternalServerError)
		return
	}

	studentDocRef := doc.Ref

	// Access the 'Goals' subcollection
	goalsCollection := studentDocRef.Collection("Goals")

	// Set up Google Sheets client
	sheetsService, err := sheets.NewService(ctx)
	if err != nil {
		log.Printf("Failed to create Sheets client: %v", err)
		http.Error(w, "Failed to create Sheets client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ID of the Google Sheet containing the schools data
	sheetID := "1HtYJv-dcDYogeDs-Y9Pq_2qSsojpU9TMkcgKkZNaN6U"

	// Read the sheet data once and build a map of school names to their data
	readRange := "University Goals List!A3:H"
	resp, err := sheetsService.Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		log.Printf("Failed to read from Sheets: %v", err)
		http.Error(w, "Failed to read from Sheets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build a map from school name to the data
	schoolDataMap := make(map[string]map[string]interface{})

	for _, row := range resp.Values {
		if len(row) == 0 {
			continue
		}
		name, ok := row[0].(string)
		if !ok || name == "" {
			continue
		}

		// Columns B,C,D: ACT 25th,50th,75th percentiles
		// Columns F,G,H: SAT 25th,50th,75th percentiles

		// Initialize the percentiles slices with nil values
		actPercentiles := []interface{}{nil, nil, nil}
		satPercentiles := []interface{}{nil, nil, nil}

		// Handle ACT percentiles
		for i := 1; i <= 3; i++ {
			if len(row) > i {
				actPercentileValue := row[i]
				actPercentiles[i-1] = actPercentileValue
			}
		}

		// Handle SAT percentiles
		for i := 5; i <= 7; i++ {
			if len(row) > i {
				satPercentileValue := row[i]
				satPercentiles[i-5] = satPercentileValue
			}
		}

		schoolDataMap[name] = map[string]interface{}{
			"university":      name,
			"ACT_percentiles": actPercentiles,
			"SAT_percentiles": satPercentiles,
		}
	}

	// For each school in payload.Goals
	for _, schoolName := range payload.Goals {
		// Lookup the school's data in the schoolDataMap
		schoolData, exists := schoolDataMap[schoolName]
		if !exists {
			log.Printf("School %s not found in the sheet", schoolName)
			continue // Or handle error as needed
		}

		// Create or update the document in 'Goals' subcollection
		docRef := goalsCollection.Doc(schoolName)

		_, err = docRef.Set(ctx, schoolData)
		if err != nil {
			log.Printf("Failed to write to Firestore for school %s: %v", schoolName, err)
			continue // Or handle error as needed
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}
