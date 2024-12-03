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
		DriveURL    string `firestore:"drive_url"`
		ClassroomID string `firestore:"classroom_id"`
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

// ProcessStudents iterates through the 'students' collection and updates classroom_id
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

		// Check if 'classroom_id' is already set
		if student.Business.ClassroomID != "" {
			log.Printf("Student ID=%s already has 'classroom_id' set to '%s'. Skipping.", doc.Ref.ID, student.Business.ClassroomID)
			skipped++
			continue
		}

		log.Printf("Processing student %d: ID=%s", count, doc.Ref.ID)

		driveURL := student.Business.DriveURL
		if driveURL == "" {
			log.Printf("No 'drive_url' found for student ID=%s. Skipping.", doc.Ref.ID)
			continue
		}

		// The driveURL is the spreadsheet ID
		spreadsheetID := driveURL

		// Read cell A1 from 'data' tab with retry
		var classroomID string
		err = retry(5, 1*time.Second, func() error {
			var innerErr error
			classroomID, innerErr = readClassroomID(ctx, sheetsService, spreadsheetID)
			return innerErr
		})
		if err != nil {
			log.Printf("Error reading classroom ID for student ID=%s after retries: %v", doc.Ref.ID, err)
			continue
		}

		if classroomID == "" {
			log.Printf("No value found in 'data!A1' for student ID=%s. Skipping update.", doc.Ref.ID)
			continue
		}

		// Update the 'classroom_id' field in Firestore
		err = fu.UpdateClassroomID(ctx, doc.Ref, classroomID)
		if err != nil {
			log.Printf("Error updating 'classroom_id' for student ID=%s: %v", doc.Ref.ID, err)
			continue
		}

		log.Printf("Successfully updated 'classroom_id' for student ID=%s to '%s'.", doc.Ref.ID, classroomID)
	}

	log.Printf("Processed %d students. Skipped %d students who already had 'classroom_id' set.", count, skipped)
	return nil
}

// readClassroomID reads the value from cell A1 in the 'data' tab of the spreadsheet
func readClassroomID(ctx context.Context, sheetsService *sheets.Service, spreadsheetID string) (string, error) {
	readRange := "data!A1"
	resp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve data from '%s': %v", readRange, err)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		log.Printf("No data found in '%s'.", readRange)
		return "", nil
	}

	// Assuming A1 contains the classroom_id
	classroomID, ok := resp.Values[0][0].(string)
	if !ok {
		// If not a string, try to convert to string
		classroomID = fmt.Sprintf("%v", resp.Values[0][0])
	}

	return classroomID, nil
}

// UpdateClassroomID updates the 'business.classroom_id' field in the student document
func (fu *FirestoreUpdater) UpdateClassroomID(ctx context.Context, docRef *firestore.DocumentRef, classroomID string) error {
	_, err := docRef.Update(ctx, []firestore.Update{
		{
			Path:  "business.classroom_id",
			Value: classroomID,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update 'classroom_id': %v", err)
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
