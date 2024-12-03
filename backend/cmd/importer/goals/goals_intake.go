package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"google.golang.org/api/drive/v3"
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
		DriveURL    string `firestore:"drive_url"`
		ClassroomID string `firestore:"classroom_id"`
	} `firestore:"business"`
}

// retry executes a function up to 'maxRetries' times with exponential backoff
func retry(maxRetries int, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if the error is rate limit exceeded or a transient error
		if isRetryableError(err) {
			log.Printf("Transient error encountered: %v. Retrying in %v...", err, delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
			continue
		}

		// For non-retryable errors, do not retry
		return err
	}
	return fmt.Errorf("max retries exceeded")
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	// Example checks; adjust based on actual error messages or types
	errorMessage := strings.ToLower(err.Error())
	return strings.Contains(errorMessage, "rate limit exceeded") ||
		strings.Contains(errorMessage, "quota exceeded") ||
		strings.Contains(errorMessage, "temporarily unavailable") ||
		strings.Contains(errorMessage, "timeout") ||
		strings.Contains(errorMessage, "internal server error")
}

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

	// Initialize Context
	ctx := context.Background()

	// Initialize Google Sheets API client
	sheetService, err := sheets.NewService(ctx, option.WithCredentialsFile(serviceAccountPath), option.WithScopes(
		sheets.SpreadsheetsReadonlyScope,
	))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	// Initialize Google Drive API client (if needed)
	driveService, err := drive.NewService(ctx, option.WithCredentialsFile(serviceAccountPath), option.WithScopes(
		drive.DriveReadonlyScope,
	))
	if err != nil {
		log.Fatalf("Unable to create Drive client: %v", err)
	}

	// Initialize Firestore Client
	firestoreClient, err := firestore.NewClient(ctx, firestoreProjectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Initialize FirestoreUpdater
	fu := FirestoreUpdater{Client: firestoreClient}

	// Process students
	err = fu.ProcessStudents(ctx, sheetService, driveService)
	if err != nil {
		log.Fatalf("Error processing students: %v", err)
	}

	log.Println("Goals intake process completed.")
}

// ProcessStudents iterates through the 'students' collection and processes their Goals
func (fu *FirestoreUpdater) ProcessStudents(ctx context.Context, sheetsService *sheets.Service, driveService *drive.Service) error {
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
			if err == iterator.Done {
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

		log.Printf("Processing student %d: ID=%s, Name=%s", count, doc.Ref.ID, student.Personal.Name)

		driveURL := student.Business.DriveURL
		if driveURL == "" {
			log.Printf("No 'drive_url' found for student ID=%s. Skipping.", doc.Ref.ID)
			skipped++
			continue
		}

		// Since 'drive_url' is the Spreadsheet ID, use it directly
		spreadsheetID := driveURL

		// Read data from the student's 'Goals' sheet with retry
		var goalsDataRows []map[string]interface{}
		err = retry(5, 1*time.Second, func() error {
			var innerErr error
			goalsDataRows, innerErr = fetchGoalsSheetData(sheetsService, spreadsheetID)
			return innerErr
		})
		if err != nil {
			log.Printf("Error reading Goals sheet for student ID=%s after retries: %v", doc.Ref.ID, err)
			continue
		}
		if len(goalsDataRows) == 0 {
			log.Printf("No data found in Goals sheet for student ID=%s", doc.Ref.ID)
			continue
		}

		// Process and store each row of goals into Firestore with retry
		for _, goalData := range goalsDataRows {
			collegeName, ok := goalData["College"].(string)
			if !ok || collegeName == "" {
				continue
			}

			// Sanitize the college name to create a valid document ID
			sanitizedCollegeName := sanitizeForFirestoreID(collegeName)

			goalsCollection := doc.Ref.Collection("Goals")
			goalDocRef := goalsCollection.Doc(sanitizedCollegeName)

			// Define the write operation as a function to be retried
			writeOperation := func() error {
				_, err := goalDocRef.Set(ctx, goalData)
				if err != nil {
					return fmt.Errorf("failed to set goal document: %v", err)
				}
				return nil
			}

			// Execute the write operation with retry
			err = retry(5, 1*time.Second, writeOperation)
			if err != nil {
				log.Printf("Failed to write goal data for student ID=%s, college %s after retries: %v", doc.Ref.ID, collegeName, err)
				continue
			}

			log.Printf("Successfully wrote goal data for student ID=%s, college %s", doc.Ref.ID, collegeName)
		}
	}

	log.Printf("Processed %d students. Skipped %d students.", count, skipped)
	return nil
}

// fetchGoalsSheetData reads data from a student's 'Goals' sheet in Google Sheets
func fetchGoalsSheetData(sheetService *sheets.Service, sheetID string) ([]map[string]interface{}, error) {
	var goalsDataRows []map[string]interface{}

	// Read data from the 'Goals' sheet
	readRange := "Goals!A:G" // Adjusted to read columns A to G
	resp, err := sheetService.Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from 'Goals' tab: %v", err)
	}

	// Process each row
	for i, row := range resp.Values {
		// Skipping header rows if any
		if i < 2 {
			continue
		}

		// Helper functions to safely get cell values
		getCellStringValue := func(row []interface{}, index int) string {
			if index >= len(row) {
				return ""
			}
			if row[index] == nil {
				return ""
			}
			return strings.TrimSpace(fmt.Sprintf("%v", row[index]))
		}

		getCellFloatValue := func(row []interface{}, index int) (float64, bool) {
			if index >= len(row) {
				return 0, false
			}
			if row[index] == nil {
				return 0, false
			}
			valStr := strings.TrimSpace(fmt.Sprintf("%v", row[index]))
			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return 0, false
			}
			return val, true
		}

		college := getCellStringValue(row, 0)
		if college == "" {
			continue
		}

		// Check if this row is a header row (again)
		firstCellLower := strings.ToLower(college)
		if firstCellLower == "school" || firstCellLower == "college" {
			// Header row detected, skip it
			log.Printf("Skipping header row in data at college '%s'", college)
			continue
		}

		// Extract scores from columns B-G
		// Create slices to hold scores from columns B-D and E-G
		scores1 := []float64{}
		scores2 := []float64{}

		for idx := 1; idx <= 3; idx++ {
			val, ok := getCellFloatValue(row, idx)
			if ok {
				scores1 = append(scores1, val)
			} else {
				scores1 = append(scores1, 0)
			}
		}

		for idx := 4; idx <= 6; idx++ {
			val, ok := getCellFloatValue(row, idx)
			if ok {
				scores2 = append(scores2, val)
			} else {
				scores2 = append(scores2, 0)
			}
		}

		// Now determine whether scores1 and scores2 are ACT or SAT
		isACTScores := func(scores []float64) bool {
			countValid := 0
			for _, score := range scores {
				if score >= 1 && score <= 36 {
					countValid++
				}
			}
			return countValid >= 2 // At least 2 valid scores
		}

		isSATScores := func(scores []float64) bool {
			countValid := 0
			for _, score := range scores {
				if score >= 400 && score <= 1600 {
					countValid++
				}
			}
			return countValid >= 2 // At least 2 valid scores
		}

		data := map[string]interface{}{
			"College": college,
		}

		// Initialize empty slices for ACT and SAT percentiles
		var ACT_percentiles []float64
		var SAT_percentiles []float64

		// Process scores1
		if isACTScores(scores1) {
			ACT_percentiles = scores1
		} else if isSATScores(scores1) {
			SAT_percentiles = scores1
		}

		// Process scores2
		if isACTScores(scores2) {
			if len(ACT_percentiles) == 0 {
				ACT_percentiles = scores2
			} else {
				log.Printf("Duplicate ACT scores for college '%s', keeping first set", college)
			}
		} else if isSATScores(scores2) {
			if len(SAT_percentiles) == 0 {
				SAT_percentiles = scores2
			} else {
				log.Printf("Duplicate SAT scores for college '%s', keeping first set", college)
			}
		}

		// Add percentiles to data if they exist
		if len(ACT_percentiles) > 0 {
			data["ACT_percentiles"] = ACT_percentiles
		}
		if len(SAT_percentiles) > 0 {
			data["SAT_percentiles"] = SAT_percentiles
		}

		if len(ACT_percentiles) == 0 && len(SAT_percentiles) == 0 {
			log.Printf("No valid ACT or SAT scores found for college '%s', skipping", college)
			continue
		}

		goalsDataRows = append(goalsDataRows, data)
	}

	return goalsDataRows, nil
}

// sanitizeForFirestoreID sanitizes a string to be a valid Firestore document ID
func sanitizeForFirestoreID(input string) string {
	// Remove leading/trailing whitespace
	input = strings.TrimSpace(input)

	// Replace forbidden characters with '_'
	forbiddenChars := regexp.MustCompile(`[*/[\]#%.]`)
	sanitized := forbiddenChars.ReplaceAllString(input, "_")

	// Replace forward slashes '/' with '_'
	sanitized = strings.ReplaceAll(sanitized, "/", "_")

	// Truncate to max length if necessary
	maxLength := 1500 // Firestore document ID max length is 1500 bytes
	if len(sanitized) > maxLength {
		sanitized = sanitized[:maxLength]
	}

	return sanitized
}
