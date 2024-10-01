package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve environment variables
	serviceAccountPath := os.Getenv("SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Fatal("SERVICE_ACCOUNT_PATH is not set in the environment variables")
	}

	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	if spreadsheetID == "" {
		log.Fatal("SPREADSHEET_ID is not set in the environment variables")
	}

	readRange := os.Getenv("READ_RANGE")
	if readRange == "" {
		log.Fatal("READ_RANGE is not set in the environment variables")
	}

	firestoreProjectID := os.Getenv("FIRESTORE_PROJECT_ID")
	if firestoreProjectID == "" {
		log.Fatal("FIRESTORE_PROJECT_ID is not set in the environment variables")
	}

	// Context used for API calls
	ctx := context.Background()

	// Set Up Google Sheets API client
	sheetService, err := sheets.NewService(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	// Log the range being read
	log.Printf("Reading data from spreadsheet ID: %s, Range: %s", spreadsheetID, readRange)

	// Read data from spreadsheet
	response, err := sheetService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to read data from sheet: %v", err)
	}

	// Initialize Firestore Client
	firestoreClient, err := firestore.NewClient(ctx, firestoreProjectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close() // Ensures cleanup of resources

	// Initialize a counter for processed rows
	rowNumber := 3 // Starting from row 3

	// Iterate over the rows to process each student's data
	for _, row := range response.Values {
		rowNumber++ // Increment row number for logging

		// Extract Data from Row Safely
		studentName, ok := getCellStringValue(row, 0)
		if !ok || studentName == "" {
			log.Printf("Row %d skipped: invalid or empty student name: %v", rowNumber, row)
			continue
		}

		// Extract other columns, initializing to empty strings if missing
		status, ok := getCellStringValue(row, 1)
		if !ok {
			status = "" // Default value
			log.Printf("Row %d: Status missing for student %s, initializing to empty string.", rowNumber, studentName)
		}

		team, ok := getCellStringValue(row, 2)
		if !ok {
			team = "" // Default value
			log.Printf("Row %d: Team missing for student %s, initializing to empty string.", rowNumber, studentName)
		}

		// Column E (index 3) is skipped as per your original code

		teamLeader, ok := getCellStringValue(row, 4)
		if !ok {
			teamLeader = "" // Default value
			log.Printf("Row %d: Team Leader missing for student %s, initializing to empty string.", rowNumber, studentName)
		}

		// Pull Associated Tutors Array
		associatedTutors := parseTutors(team)

		// Check if this student exists in DB and get the current document data
		docRef, existingData, err := getStudentData(firestoreClient, ctx, studentName)
		if err != nil {
			log.Printf("Row %d: Error checking if student %s exists: %v", rowNumber, studentName, err)
			continue
		}

		// Prepare new data map for 'business' subdocument
		businessData := map[string]interface{}{
			"status":            status,
			"associated_tutors": associatedTutors,
			"team_lead":         teamLeader,
			"remaining_hours":   0, // Initialize 'remaining_hours' to 0
		}

		// Prepare new data map for 'personal' subdocument
		personalData := map[string]interface{}{
			"name": studentName,
		}

		if existingData != nil {
			// Check if new data differs from existing data
			if !dataEquals(existingData, businessData) {
				// Update Firestore with new data in the 'business' and 'personal' subdocuments
				err = updateSubdocument(firestoreClient, ctx, docRef, "business", businessData)
				if err != nil {
					log.Printf("Row %d: Failed to update business data for student %s: %v", rowNumber, studentName, err)
					continue
				}
				err = updateSubdocument(firestoreClient, ctx, docRef, "personal", personalData)
				if err != nil {
					log.Printf("Row %d: Failed to update personal data for student %s: %v", rowNumber, studentName, err)
					continue
				}
				log.Printf("Row %d: Updated student: %s", rowNumber, studentName)
			} else {
				log.Printf("Row %d: Student %s data is unchanged, skipping update.", rowNumber, studentName)
			}
		} else {
			// Student does not exist, create new document and subdocuments
			studentID := uuid.New().String()
			docRef := firestoreClient.Collection("students").Doc(studentID)

			err = updateSubdocument(firestoreClient, ctx, docRef, "business", businessData)
			if err != nil {
				log.Printf("Row %d: Failed to add business data for student %s: %v", rowNumber, studentName, err)
				continue
			}

			err = updateSubdocument(firestoreClient, ctx, docRef, "personal", personalData)
			if err != nil {
				log.Printf("Row %d: Failed to add personal data for student %s: %v", rowNumber, studentName, err)
				continue
			}

			log.Printf("Row %d: Successfully added student: %s", rowNumber, studentName)
		}
	}

	log.Println("Intake process completed.")
}

// Helper function to update a subdocument for a student document
func updateSubdocument(firestoreClient *firestore.Client, ctx context.Context, docRef *firestore.DocumentRef, subDocName string, data map[string]interface{}) error {
	_, err := docRef.Set(ctx, map[string]interface{}{subDocName: data}, firestore.MergeAll)
	if err != nil {
		log.Printf("Failed to update subdocument %s: %v", subDocName, err)
		return err
	}
	return nil
}

// Helper - Parse Tutors from the 'team' column.
func parseTutors(team string) []string {
	// Split passed string by a '/' or by ' only'
	team = strings.ReplaceAll(team, " only", "")
	tutors := strings.Split(team, "/")

	for i, tutor := range tutors {
		tutors[i] = strings.TrimSpace(tutor)
	}

	return tutors
}

// Helper - Check if student with the given name already exists and return the document reference and data if so.
func getStudentData(client *firestore.Client, ctx context.Context, name string) (*firestore.DocumentRef, map[string]interface{}, error) {
	// Query based on the 'personal.name' field
	iter := client.Collection("students").Where("personal.name", "==", name).Limit(1).Documents(ctx)
	defer iter.Stop() // Ensures cleanup of resources

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, nil, nil // No document found
		}
		return nil, nil, err // Error during iteration.
	}

	return doc.Ref, doc.Data(), nil
}

// Helper - Compare existing data with new data to determine if an update is needed.
func dataEquals(existingData map[string]interface{}, businessData map[string]interface{}) bool {
	// Convert associated_tutors from []interface{} to []string
	existingTutors, ok := existingData["associated_tutors"].([]interface{})
	if !ok {
		return false // If existing data is not in the expected format, consider it as a mismatch
	}

	// Convert []interface{} to []string
	var existingTutorsStr []string
	for _, tutor := range existingTutors {
		tutorStr, ok := tutor.(string)
		if !ok {
			return false // If conversion fails, treat it as a mismatch
		}
		existingTutorsStr = append(existingTutorsStr, tutorStr)
	}

	// Compare relevant fields
	newTutors, ok := businessData["associated_tutors"].([]string)
	if !ok {
		return false
	}

	return existingData["status"] == businessData["status"] &&
		strings.Join(existingTutorsStr, ",") == strings.Join(newTutors, ",") &&
		existingData["team_lead"] == businessData["team_lead"] &&
		existingData["remaining_hours"] == businessData["remaining_hours"]
}

// Helper function to safely extract string values from a cell
func getCellStringValue(row []interface{}, index int) (string, bool) {
	if index >= len(row) {
		return "", false
	}
	if row[index] == nil {
		return "", false
	}

	switch v := row[index].(type) {
	case string:
		return strings.TrimSpace(v), true
	case float64:
		// If the cell contains a number but expected a string, convert it to string
		return fmt.Sprintf("%.0f", v), true
	default:
		return "", false
	}
}
