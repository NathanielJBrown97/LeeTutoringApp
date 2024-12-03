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

	// Process Test Data
	err = fu.ProcessTestData(ctx, sheetsService)
	if err != nil {
		log.Fatalf("Error processing Test Data: %v", err)
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

// ProcessTestData iterates through the 'students' collection and updates Test Data subcollection
func (fu *FirestoreUpdater) ProcessTestData(ctx context.Context, sheetsService *sheets.Service) error {
	// Reference to the 'students' collection
	studentsCollection := fu.Client.Collection("students")

	// Define allowed test types
	allowedTestTypes := map[string]bool{
		"Practice":      true,
		"Official":      true,
		"Unofficial SS": true,
		"Official SS":   true,
	}

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

		driveURL := student.Business.DriveURL
		if driveURL == "" {
			log.Printf("No 'drive_url' found for student ID=%s. Skipping.", doc.Ref.ID)
			skipped++
			continue
		}

		// The driveURL is the spreadsheet ID
		spreadsheetID := driveURL

		// Read 'Test Data' sheet with retry
		rows, err := retryTestData(5, 1*time.Second, func() ([][]interface{}, error) {
			return readTestData(ctx, sheetsService, spreadsheetID)
		})
		if err != nil {
			log.Printf("Error reading Test Data for student ID=%s after retries: %v", doc.Ref.ID, err)
			continue
		}

		if len(rows) < 2 {
			log.Printf("No data rows found in 'Test Data' for student ID=%s. Skipping.", doc.Ref.ID)
			continue
		}

		// Process each row in the Test Data sheet, skipping the header row
		for i, row := range rows {
			if i == 0 {
				// Skip header row
				continue
			}

			if len(row) < 14 { // Ensure there are at least 14 columns (A to N)
				log.Printf("Row %d has insufficient columns for student ID=%s. Skipping row.", i+1, doc.Ref.ID)
				continue
			}

			// Extract and process necessary columns
			baseline := parseBoolean(row[0])
			typeStr := parseString(row[1])
			testStr := parseString(row[2])
			dateRaw := parseString(row[3])
			formattedDate := strings.ReplaceAll(dateRaw, "/", "-")

			// Validate the test type
			if !allowedTestTypes[typeStr] {
				log.Printf("Invalid test type '%s' for student ID=%s. Skipping row.", typeStr, doc.Ref.ID)
				continue
			}

			// Document ID: Combine B + C + D with spaces, date formatted
			docID := fmt.Sprintf("%s %s %s", typeStr, testStr, formattedDate)

			// Prepare ACT_Scores and SAT_Scores maps
			actScores := map[string]interface{}{
				"English":   parseNumber(row[8]),
				"Math":      parseNumber(row[9]),
				"Reading":   parseNumber(row[10]),
				"Science":   parseNumber(row[11]),
				"ACT_Total": parseNumber(row[13]),
			}

			satScores := map[string]interface{}{
				"EBRW":      parseNumber(row[4]),
				"Math":      parseNumber(row[5]),
				"Reading":   parseNumber(row[6]),
				"Writing":   parseNumber(row[7]),
				"SAT_Total": parseNumber(row[12]),
			}

			// Prepare data for the main Test Data document
			mainDocData := map[string]interface{}{
				"baseline":   baseline,
				"type":       typeStr,
				"test":       testStr,
				"date":       dateRaw,
				"ACT_Scores": actScores,
				"SAT_Scores": satScores,
				"updated_at": time.Now(), // Optional: Timestamp for when the data was updated
			}

			// Reference to the 'Test Data' subcollection
			testDataCollection := doc.Ref.Collection("Test Data")

			// Create or update the main Test Data document
			testDataDocRef := testDataCollection.Doc(docID)
			_, err = testDataDocRef.Set(ctx, mainDocData)
			if err != nil {
				log.Printf("Error writing Test Data for student ID=%s, docID=%s: %v", doc.Ref.ID, docID, err)
				continue
			}

			log.Printf("Successfully wrote Test Data for student ID=%s, docID=%s.", doc.Ref.ID, docID)
		}
	}

	log.Printf("Processed %d students. Skipped %d students due to missing 'drive_url'.", count, skipped)
	return nil
}

// readTestData reads all rows from the 'Test Data' tab of the spreadsheet
func readTestData(ctx context.Context, sheetsService *sheets.Service, spreadsheetID string) ([][]interface{}, error) {
	readRange := "Test Data"
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

// retryTestData executes a function up to 'maxRetries' times with exponential backoff for Test Data
func retryTestData(maxRetries int, initialDelay time.Duration, fn func() ([][]interface{}, error)) ([][]interface{}, error) {
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

// parseBoolean attempts to parse an interface{} to a boolean. Returns false if parsing fails.
func parseBoolean(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "yes" || lower == "1" {
			return true
		}
		return false
	case float64:
		return v != 0
	default:
		return false
	}
}

// parseString attempts to parse an interface{} to a string. Returns empty string if parsing fails.
func parseString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case float64:
		// If the value is a number but expected as string, format without decimal if possible
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// parseNumber attempts to parse an interface{} to a float64 pointer. Returns nil if parsing fails.
func parseNumber(value interface{}) *float64 {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case float64:
		return &v
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || trimmed == "-" {
			return nil
		}
		var num float64
		_, err := fmt.Sscanf(trimmed, "%f", &num)
		if err != nil {
			return nil
		}
		return &num
	default:
		return nil
	}
}
