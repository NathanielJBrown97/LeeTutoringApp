package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// FirestoreUpdater handles Firestore operations
type FirestoreUpdater struct {
	Client *firestore.Client
}

// StudentDocument represents the structure of a student document in Firestore
type StudentDocument struct {
	Personal struct {
		Name string `firestore:"name"`
	} `firestore:"personal"`
	Business struct {
		Status           string   `firestore:"status"`
		AssociatedTutors []string `firestore:"associated_tutors"`
		Notes            string   `firestore:"notes"`
		TeamLead         string   `firestore:"team_lead"`
	} `firestore:"business"`
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	projectID := os.Getenv("FIRESTORE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("FIRESTORE_PROJECT_ID is not set in the .env file")
	}

	serviceAccountPath := os.Getenv("SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Fatal("SERVICE_ACCOUNT_PATH is not set in the .env file")
	}

	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	if spreadsheetID == "" {
		log.Fatal("SPREADSHEET_ID is not set in the .env file")
	}

	sheetName := os.Getenv("SHEET_NAME")
	if sheetName == "" {
		log.Fatal("SHEET_NAME is not set in the .env file")
	}

	// Initialize Firestore client
	ctx := context.Background()
	firestoreClient, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer func() {
		if err := firestoreClient.Close(); err != nil {
			log.Printf("Error closing Firestore client: %v", err)
		}
	}()

	// Initialize Sheets service
	sheetsService, err := initializeSheetsService(serviceAccountPath)
	if err != nil {
		log.Fatalf("Failed to initialize Sheets service: %v", err)
	}

	// Initialize FirestoreUpdater
	fu := FirestoreUpdater{Client: firestoreClient}

	// Process students
	err = fu.ProcessStudents(ctx, sheetsService, spreadsheetID, sheetName)
	if err != nil {
		log.Fatalf("Error processing students: %v", err)
	}

	log.Println("All students have been processed successfully.")
}

// initializeSheetsService initializes the Google Sheets service
func initializeSheetsService(serviceAccountPath string) (*sheets.Service, error) {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return nil, fmt.Errorf("unable to create Sheets client: %v", err)
	}
	return sheetsService, nil
}

// ProcessStudents iterates through the 'students' collection and updates relevant fields
func (fu *FirestoreUpdater) ProcessStudents(ctx context.Context, sheetsService *sheets.Service, spreadsheetID, sheetName string) error {
	// Reference to the 'students' collection
	studentsCollection := fu.Client.Collection("students")

	// Retrieve all student documents
	iter := studentsCollection.Documents(ctx)
	defer iter.Stop()

	count := 0
	skipped := 0
	updated := 0

	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done { // Corrected error check
				break
			}
			return fmt.Errorf("error iterating through students: %v", err)
		}

		count++

		// Parse the student document into StudentDocument struct
		var student StudentDocument
		err = doc.DataTo(&student)
		if err != nil {
			log.Printf("Error parsing student document ID=%s: %v", doc.Ref.ID, err)
			continue
		}

		studentName := strings.TrimSpace(student.Personal.Name)
		if studentName == "" {
			log.Printf("Student ID=%s has an empty name in 'personal' subdocument. Skipping.", doc.Ref.ID)
			skipped++
			continue
		}

		log.Printf("Processing student %d: ID=%s, Name=%s", count, doc.Ref.ID, studentName)

		// Find the row in the spreadsheet that matches the student's name in column B
		var row []interface{}
		err = retry(5, 1*time.Second, func() error {
			var innerErr error
			row, innerErr = findRowByName(ctx, sheetsService, spreadsheetID, sheetName, studentName)
			return innerErr
		})
		if err != nil {
			log.Printf("Error finding row for student ID=%s, Name=%s after retries: %v", doc.Ref.ID, studentName, err)
			continue
		}

		if row == nil {
			log.Printf("No matching row found in spreadsheet for student ID=%s, Name=%s. Skipping.", doc.Ref.ID, studentName)
			skipped++
			continue
		}

		// Extract data from the row
		status, _ := getStringFromRow(row, 2)              // Column C (index 2)
		associatedTutorsStr, _ := getStringFromRow(row, 3) // Column D (index 3)
		notes, _ := getStringFromRow(row, 4)               // Column E (index 4)
		teamLead, _ := getStringFromRow(row, 5)            // Column F (index 5)

		// Process associated_tutors
		associatedTutors := processAssociatedTutors(associatedTutorsStr)

		// Prepare updates
		updates := []firestore.Update{
			{
				Path:  "business.status",
				Value: status,
			},
			{
				Path:  "business.associated_tutors",
				Value: associatedTutors,
			},
			{
				Path:  "business.notes",
				Value: notes,
			},
			{
				Path:  "business.team_lead",
				Value: teamLead,
			},
		}

		// Update the Firestore document
		err = fu.UpdateBusinessFields(ctx, doc.Ref, updates)
		if err != nil {
			log.Printf("Error updating business fields for student ID=%s: %v", doc.Ref.ID, err)
			continue
		}

		log.Printf("Successfully updated business fields for student ID=%s.", doc.Ref.ID)
		updated++
	}

	log.Printf("Processed %d students. Skipped %d students. Updated %d students.", count, skipped, updated)
	return nil
}

// findRowByName searches for the student's name in column B and returns the entire row if found
func findRowByName(ctx context.Context, sheetsService *sheets.Service, spreadsheetID, sheetName, name string) ([]interface{}, error) {
	readRange := fmt.Sprintf("%s!B:B", sheetName) // e.g., "Current Students!B:B"

	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from '%s': %v", readRange, err)
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("no data found in column B")
	}

	// Normalize the name for matching
	normalizedTargetName := normalizeName(name)

	for i, row := range resp.Values {
		if len(row) == 0 {
			continue
		}
		cellValue, ok := row[0].(string)
		if !ok {
			// If not a string, convert to string
			cellValue = fmt.Sprintf("%v", row[0])
		}
		normalizedCellValue := normalizeName(cellValue)
		if normalizedCellValue == normalizedTargetName {
			// Fetch the entire row (columns A to F)
			rowRange := fmt.Sprintf("%s!A%d:F%d", sheetName, i+1, i+1) // Columns A to F
			rowResp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, rowRange).Do()
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve row data for '%s': %v", name, err)
			}
			if len(rowResp.Values) == 0 {
				return nil, fmt.Errorf("no data found in the matched row for '%s'", name)
			}
			return rowResp.Values[0], nil
		}
	}

	return nil, nil // No match found
}

// normalizeName trims spaces and converts the name to lowercase
func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// getStringFromRow safely retrieves a string from a specific column in a row
func getStringFromRow(row []interface{}, index int) (string, error) {
	if index >= len(row) {
		return "", nil // No data in this column
	}
	value, ok := row[index].(string)
	if !ok {
		// Attempt to convert non-string types to string
		value = fmt.Sprintf("%v", row[index])
	}
	return strings.TrimSpace(value), nil
}

// processAssociatedTutors processes the associated_tutors string into a normalized array
func processAssociatedTutors(tutorsStr string) []string {
	if tutorsStr == "" {
		return []string{}
	}

	tutorsStr = strings.ToLower(tutorsStr)
	tutorsStr = strings.TrimSpace(tutorsStr)

	// Replace " only" with empty string and trim
	tutorsStr = strings.ReplaceAll(tutorsStr, " only", "")
	tutorsStr = strings.ReplaceAll(tutorsStr, "only", "")
	tutorsStr = strings.TrimSpace(tutorsStr)

	// Split by '/'
	tutorNames := strings.Split(tutorsStr, "/")

	var associatedTutors []string
	for _, tutor := range tutorNames {
		tutor = strings.TrimSpace(tutor)
		if tutor == "" {
			continue
		}

		// Capitalize the first letter of each word
		tutor = capitalizeWords(tutor)

		// Handle special cases
		if tutor == "Not Started Yet" || tutor == "Work In Progress" {
			associatedTutors = append(associatedTutors, tutor)
			continue
		}

		associatedTutors = append(associatedTutors, tutor)
	}

	return associatedTutors
}

// capitalizeWords capitalizes the first letter of each word in a string
func capitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		// Handle cases like "Edward/Kyra" which may have been split by '/'
		words[i] = strings.Title(word)
	}
	return strings.Join(words, " ")
}

// UpdateBusinessFields updates multiple fields in the 'business' subdocument
func (fu *FirestoreUpdater) UpdateBusinessFields(ctx context.Context, docRef *firestore.DocumentRef, updates []firestore.Update) error {
	_, err := docRef.Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update business fields: %v", err)
	}
	return nil
}

// retry executes a function up to 'maxRetries' times with exponential backoff
func retry(maxRetries int, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if the error is rate limit exceeded
		if isRateLimitError(err) {
			log.Printf("Rate limit exceeded. Retrying in %v...", delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
			continue
		}

		// For other errors, do not retry
		return err
	}
	return fmt.Errorf("max retries exceeded")
}

// isRateLimitError checks if the error is a rate limit error
func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "RATE_LIMIT_EXCEEDED")
}
