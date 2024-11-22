package firestoreupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// UpdateTestDatesHandler handles the POST request from the GAS trigger
func (app *App) UpdateTestDatesHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the JSON payload
	var payload struct {
		SpreadsheetID string `json:"spreadsheetId"`
		SheetName     string `json:"sheetName"`
		UserEmail     string `json:"userEmail"`
	}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received trigger from user %s for spreadsheet %s sheet %s", payload.UserEmail, payload.SpreadsheetID, payload.SheetName)

	// Process the spreadsheet and update Firestore
	err = app.processSpreadsheet(payload.SpreadsheetID, payload.SheetName)
	if err != nil {
		log.Printf("Failed to process spreadsheet: %v", err)
		http.Error(w, "Failed to process spreadsheet", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processSpreadsheet reads the 'Test Dates' sheet and updates Firestore
func (app *App) processSpreadsheet(spreadsheetID string, sheetName string) error {
	ctx := context.Background()

	// Initialize the Sheets API client
	sheetsService, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	// Read data from the 'Test Dates' tab
	readRange := fmt.Sprintf("%s!A2:E", sheetName)
	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	// Check if any data was retrieved
	if len(resp.Values) == 0 {
		log.Println("No data found in the 'Test Dates' tab.")
		return nil
	}

	// Iterate over each row and update Firestore
	for _, row := range resp.Values {
		// Ensure the row has enough columns
		if len(row) < 5 {
			log.Printf("Skipping incomplete row: %v", row)
			continue
		}

		testType := fmt.Sprintf("%v", row[0])
		testDate := fmt.Sprintf("%v", row[1])
		sanitizedDate := strings.ReplaceAll(testDate, "/", "-")
		regDeadline := fmt.Sprintf("%v", row[2])
		lateRegDeadline := fmt.Sprintf("%v", row[3])
		scoreReleaseDate := fmt.Sprintf("%v", row[4])

		// Construct document name and data
		docName := fmt.Sprintf("%s %s", testType, sanitizedDate)
		data := map[string]interface{}{
			"Test Type":                     testType,
			"Test Date":                     sanitizedDate,
			"Regular Registration Deadline": regDeadline,
			"Late Registration Deadline":    lateRegDeadline,
			"Score Release Date":            scoreReleaseDate,
			"Notes":                         "", // Initialize to blank or retain existing
		}

		// Iterate over each student
		iter := app.FirestoreClient.Collection("students").Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Failed to iterate students: %v", err)
				break
			}

			studentID := doc.Ref.ID

			// Reference to the student's 'Test Dates' subcollection
			subDocRef := doc.Ref.Collection("Test Dates").Doc(docName)

			// Check if the document already exists
			_, err = subDocRef.Get(ctx)
			if err != nil {
				// Document does not exist, create it
				_, err = subDocRef.Set(ctx, data)
				if err != nil {
					log.Printf("Failed to create document '%s' for student '%s': %v", docName, studentID, err)
				} else {
					log.Printf("Created document '%s' for student '%s'", docName, studentID)
				}
			} else {
				// Document exists, skip
				log.Printf("Document '%s' already exists for student '%s', skipping", docName, studentID)
			}
		}
	}

	return nil
}
