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

// TestDate represents the structure of a test date entry
type TestDate struct {
	TestType                    string `firestore:"test_type"`
	TestDate                    string `firestore:"test_date"`
	RegularRegistrationDeadline string `firestore:"regular_registration_deadline"`
	LateRegistrationDeadline    string `firestore:"late_registration_deadline"`
	ScoreReleaseDate            string `firestore:"score_release_date"`
	Notes                       string `firestore:"notes"` // New field for notes
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
	err = fu.ProcessStudents(ctx, sheetsService)
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

// ProcessStudents iterates through the 'students' collection and updates 'Test Dates' subcollection
func (fu *FirestoreUpdater) ProcessStudents(ctx context.Context, sheetsService *sheets.Service) error {
	// Reference to the 'students' collection
	studentsCollection := fu.Client.Collection("students")

	// Retrieve all student documents
	iter := studentsCollection.Documents(ctx)
	defer iter.Stop()

	count := 0
	skipped := 0
	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done { // Corrected error check
				break
			}
			return fmt.Errorf("error iterating through students: %v", err)
		}

		count++

		log.Printf("Processing student %d: ID=%s", count, doc.Ref.ID)

		// Safely extract 'drive_url' from the nested 'business' map
		businessData, ok := doc.Data()["business"].(map[string]interface{})
		if !ok {
			log.Printf("Invalid or missing 'business' field for student ID=%s. Skipping.", doc.Ref.ID)
			skipped++
			continue
		}

		driveURL, ok := businessData["drive_url"].(string)
		if !ok || driveURL == "" {
			log.Printf("No valid 'drive_url' found for student ID=%s. Skipping.", doc.Ref.ID)
			skipped++
			continue
		}

		spreadsheetID := driveURL

		// Read 'Test Dates' sheet with retry
		testDates, err := retryTestDates(ctx, sheetsService, spreadsheetID)
		if err != nil {
			log.Printf("Error reading 'Test Dates' for student ID=%s after retries: %v", doc.Ref.ID, err)
			skipped++
			continue
		}

		if len(testDates) == 0 {
			log.Printf("No 'Test Dates' data found for student ID=%s. Skipping.", doc.Ref.ID)
			skipped++
			continue
		}

		// Create or update 'Test Dates' subcollection
		err = fu.UpdateTestDates(ctx, doc.Ref, testDates)
		if err != nil {
			log.Printf("Error updating 'Test Dates' for student ID=%s: %v", doc.Ref.ID, err)
			skipped++
			continue
		}

		log.Printf("Successfully updated 'Test Dates' for student ID=%s.", doc.Ref.ID)
	}

	log.Printf("Processed %d students. Skipped %d students due to errors or missing data.", count, skipped)
	return nil
}

// retryTestDates retries reading the 'Test Dates' sheet with exponential backoff
func retryTestDates(ctx context.Context, sheetsService *sheets.Service, spreadsheetID string) ([]TestDate, error) {
	var testDates []TestDate
	err := retry(5, 1*time.Second, func() error {
		var innerErr error
		testDates, innerErr = readTestDates(ctx, sheetsService, spreadsheetID)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return testDates, nil
}

// readTestDates reads all rows from the 'Test Dates' sheet
func readTestDates(ctx context.Context, sheetsService *sheets.Service, spreadsheetID string) ([]TestDate, error) {
	readRange := "Test Dates!A2:F" // Updated to include column F
	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from '%s': %v", readRange, err)
	}

	if len(resp.Values) == 0 {
		log.Printf("No data found in '%s'.", readRange)
		return nil, nil
	}

	var testDates []TestDate
	for i, row := range resp.Values {
		// Ensure the row has at least 5 columns (A-E). Column F is optional.
		if len(row) < 5 {
			log.Printf("Row %d has insufficient columns. Expected at least 5, got %d. Skipping.", i+2, len(row))
			continue
		}

		testType := fmt.Sprintf("%v", row[0])
		testDate := fmt.Sprintf("%v", row[1])
		regularRegDeadline := fmt.Sprintf("%v", row[2])
		lateRegDeadline := fmt.Sprintf("%v", row[3])
		scoreReleaseDate := fmt.Sprintf("%v", row[4])

		// Handle column F (Notes)
		notes := ""
		if len(row) >= 6 {
			notes = fmt.Sprintf("%v", row[5])
		}

		testDates = append(testDates, TestDate{
			TestType:                    testType,
			TestDate:                    testDate,
			RegularRegistrationDeadline: regularRegDeadline,
			LateRegistrationDeadline:    lateRegDeadline,
			ScoreReleaseDate:            scoreReleaseDate,
			Notes:                       notes, // Assign notes
		})
	}

	return testDates, nil
}

// UpdateTestDates creates or updates the 'Test Dates' subcollection for a student
func (fu *FirestoreUpdater) UpdateTestDates(ctx context.Context, docRef *firestore.DocumentRef, testDates []TestDate) error {
	// Reference to the 'Test Dates' subcollection
	testDatesCollection := docRef.Collection("Test Dates")

	for _, td := range testDates {
		// Create document ID by concatenating Test Type and Test Date with '-' instead of '/'
		sanitizedDate := strings.ReplaceAll(td.TestDate, "/", "-")
		docID := fmt.Sprintf("%s %s", td.TestType, sanitizedDate)

		// Prepare the data map
		data := map[string]interface{}{
			"test_type":                     td.TestType,
			"test_date":                     td.TestDate,
			"regular_registration_deadline": td.RegularRegistrationDeadline,
			"late_registration_deadline":    td.LateRegistrationDeadline,
			"score_release_date":            td.ScoreReleaseDate,
			"notes":                         td.Notes, // Include notes
		}

		// Set the document (creates or updates)
		_, err := testDatesCollection.Doc(docID).Set(ctx, data)
		if err != nil {
			return fmt.Errorf("failed to set Test Date document '%s': %v", docID, err)
		}
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
