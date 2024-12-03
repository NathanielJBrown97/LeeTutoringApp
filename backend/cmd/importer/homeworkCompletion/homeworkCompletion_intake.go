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
	Business struct {
		DriveURL string `firestore:"drive_url"`
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

// ProcessStudents iterates through the 'students' collection and updates Homework Completion subcollection
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

		// Parse the student document into StudentDocument struct
		var student StudentDocument
		err = doc.DataTo(&student)
		if err != nil {
			log.Printf("Error parsing student document ID=%s: %v", doc.Ref.ID, err)
			continue
		}

		driveURL := student.Business.DriveURL
		if driveURL == "" {
			log.Printf("No 'drive_url' found for student ID=%s. Skipping.", doc.Ref.ID)
			skipped++
			continue
		}

		// The driveURL is the spreadsheet ID
		spreadsheetID := driveURL

		// Read 'homework completion' sheet with retry
		rows, err := retry(5, 1*time.Second, func() (interface{}, error) {
			return readHomeworkCompletion(ctx, sheetsService, spreadsheetID)
		})
		if err != nil {
			log.Printf("Error reading homework completion for student ID=%s after retries: %v", doc.Ref.ID, err)
			continue
		}

		hwRows, ok := rows.([][]interface{})
		if !ok {
			log.Printf("Unexpected data format for student ID=%s. Skipping.", doc.Ref.ID)
			continue
		}

		// Process each row in the homework completion sheet
		for _, row := range hwRows {
			if len(row) < 2 {
				log.Printf("Row has insufficient columns for student ID=%s. Skipping row.", doc.Ref.ID)
				continue
			}

			// Extract date and percentage_complete from columns A and B
			dateRaw, ok := row[0].(string)
			if !ok {
				// If not a string, try to convert to string
				dateRaw = fmt.Sprintf("%v", row[0])
			}
			percentageComplete, ok := row[1].(string)
			if !ok {
				percentageComplete = fmt.Sprintf("%v", row[1])
			}

			// Convert date format from MM/DD/YYYY to MM-DD-YYYY
			formattedDate := strings.ReplaceAll(dateRaw, "/", "-")

			// Reference to the 'Homework Completion' subcollection
			hwCollection := doc.Ref.Collection("Homework Completion")

			// Create or update the subdocument with the formatted date as the document ID
			hwDocRef := hwCollection.Doc(formattedDate)
			_, err := hwDocRef.Set(ctx, map[string]string{
				"date":                dateRaw,
				"percentage_complete": percentageComplete,
			})
			if err != nil {
				log.Printf("Error writing Homework Completion for student ID=%s, date=%s: %v", doc.Ref.ID, formattedDate, err)
				continue
			}

			log.Printf("Successfully wrote Homework Completion for student ID=%s, date=%s.", doc.Ref.ID, formattedDate)
		}
	}

	log.Printf("Processed %d students. Skipped %d students due to missing 'drive_url'.", count, skipped)
	return nil
}

// readHomeworkCompletion reads all rows from the 'homework completion' tab of the spreadsheet
func readHomeworkCompletion(ctx context.Context, sheetsService *sheets.Service, spreadsheetID string) ([][]interface{}, error) {
	readRange := "homework completion"
	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from '%s': %v", readRange, err)
	}

	if len(resp.Values) == 0 {
		log.Printf("No data found in '%s'.", readRange)
		return [][]interface{}{}, nil
	}

	return resp.Values, nil
}

// retry executes a function up to 'maxRetries' times with exponential backoff
func retry(maxRetries int, initialDelay time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	delay := initialDelay
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		// Check if the error is rate limit exceeded
		if isRateLimitError(err) {
			log.Printf("Rate limit exceeded. Retrying in %v...", delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
			continue
		}

		// For other errors, do not retry
		return nil, err
	}
	return nil, fmt.Errorf("max retries exceeded")
}

// isRateLimitError checks if the error is a rate limit error
func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "RATE_LIMIT_EXCEEDED")
}
