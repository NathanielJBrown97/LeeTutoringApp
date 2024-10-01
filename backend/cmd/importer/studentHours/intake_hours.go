package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	// Load environment variables from the .env file located one directory up
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve environment variables
	serviceAccountPath := os.Getenv("SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Fatal("SERVICE_ACCOUNT_PATH is not set in the environment variables")
	}

	firestoreProjectID := os.Getenv("FIRESTORE_PROJECT_ID")
	if firestoreProjectID == "" {
		log.Fatal("FIRESTORE_PROJECT_ID is not set in the environment variables")
	}

	calLifeHoursSpreadsheetID := os.Getenv("CAL_LIFE_HOURS_SPREADSHEET_ID")
	if calLifeHoursSpreadsheetID == "" {
		log.Fatal("CAL_LIFE_HOURS_SPREADSHEET_ID is not set in the environment variables")
	}

	spreadsheetReadRange := os.Getenv("SPREADSHEET_READ_RANGE")
	if spreadsheetReadRange == "" {
		log.Fatal("SPREADSHEET_READ_RANGE is not set in the environment variables")
	}

	// Context used for API calls
	ctx := context.Background()

	// Set up Google Sheets API client
	sheetService, err := sheets.NewService(ctx, option.WithCredentialsFile(serviceAccountPath), option.WithScopes(
		"https://www.googleapis.com/auth/spreadsheets.readonly",
	))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	// Initialize Firestore Client
	firestoreClient, err := firestore.NewClient(ctx, firestoreProjectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close() // Ensures cleanup of resources

	// Read data from the spreadsheet
	log.Printf("Reading data from spreadsheet ID: %s, Range: %s", calLifeHoursSpreadsheetID, spreadsheetReadRange)
	response, err := sheetService.Spreadsheets.Values.Get(calLifeHoursSpreadsheetID, spreadsheetReadRange).Do()
	if err != nil {
		log.Fatalf("Unable to read data from sheet: %v", err)
	}

	// Create a map of student names to lifetime hours for quick lookup
	studentHoursMap := make(map[string]float64)
	for _, row := range response.Values {
		if len(row) < 2 {
			continue // Skip rows that don't have both name and hours
		}

		// Safely extract the name and hours
		nameCell, ok := row[0].(string)
		if !ok {
			log.Printf("Invalid name cell, skipping row: %v", row)
			continue
		}
		name := strings.TrimSpace(nameCell)

		hoursStr := ""
		switch v := row[1].(type) {
		case string:
			hoursStr = strings.TrimSpace(v)
		case float64:
			hoursStr = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			log.Printf("Invalid hours cell for student %s, skipping row: %v", name, row)
			continue
		}

		// Parse hoursStr to float64
		hours, err := strconv.ParseFloat(hoursStr, 64)
		if err != nil {
			log.Printf("Invalid hours value for student %s: %v", name, err)
			continue
		}

		studentHoursMap[name] = hours
	}

	// Iterate over all student documents in Firestore
	iter := firestoreClient.Collection("students").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error iterating through students: %v", err)
			continue
		}

		// Access the 'personal' subdocument to get the student's name
		personalData, ok := doc.Data()["personal"].(map[string]interface{})
		if !ok {
			log.Printf("Error reading 'personal' subdocument for document ID %s", doc.Ref.ID)
			continue
		}

		studentName, ok := personalData["name"].(string)
		if !ok || studentName == "" {
			log.Printf("Error reading student name from 'personal' subdocument in document ID %s", doc.Ref.ID)
			continue
		}

		// Look up the student's lifetime hours from the map
		hours, found := studentHoursMap[studentName]
		if !found {
			log.Printf("Student %s not found in the Google Sheet, skipping.", studentName)
			continue
		}

		// Prepare the data to update in the 'business' subdocument
		businessUpdate := map[string]interface{}{
			"lifetime_hours": hours,
		}

		// Update the 'business' subdocument with 'lifetime_hours'
		err = updateSubdocument(firestoreClient, ctx, doc.Ref, "business", businessUpdate)
		if err != nil {
			log.Printf("Failed to update lifetime_hours for student %s: %v", studentName, err)
			continue
		}

		log.Printf("Successfully updated lifetime_hours for student: %s", studentName)
	}

	log.Println("Data intake process completed.")
}

// Helper function to update a subdocument for a student document
func updateSubdocument(firestoreClient *firestore.Client, ctx context.Context, docRef *firestore.DocumentRef, subDocName string, data map[string]interface{}) error {
	// Use a transaction to update the subdocument
	err := firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Get the existing data
		_, err := tx.Get(docRef)
		if err != nil {
			return err
		}

		// Prepare the update data
		updates := []firestore.Update{
			{
				Path:  subDocName + ".lifetime_hours",
				Value: data["lifetime_hours"],
			},
		}

		// Update the document
		return tx.Update(docRef, updates)
	})
	if err != nil {
		log.Printf("Failed to update subdocument %s: %v", subDocName, err)
		return err
	}
	return nil
}
