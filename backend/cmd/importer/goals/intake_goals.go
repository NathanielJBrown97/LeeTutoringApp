package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Iterate over each student in Firestore before querying Google Drive
func iterateStudentsFromDB(firestoreClient *firestore.Client, ctx context.Context) ([]*firestore.DocumentSnapshot, error) {
	iter := firestoreClient.Collection("students").Documents(ctx)
	defer iter.Stop()

	var studentDocs []*firestore.DocumentSnapshot

	// Fetch each student document
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error iterating through students: %v", err)
			return nil, err
		}

		studentDocs = append(studentDocs, doc)
	}

	return studentDocs, nil
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

	studentsFolderID := os.Getenv("STUDENTS_FOLDER_ID")
	if studentsFolderID == "" {
		log.Fatal("STUDENTS_FOLDER_ID is not set in the environment variables")
	}

	// Context used for API calls
	ctx := context.Background()

	// Initialize Google Sheets API client
	sheetService, err := sheets.NewService(ctx, option.WithCredentialsFile(serviceAccountPath), option.WithScopes(
		"https://www.googleapis.com/auth/spreadsheets.readonly",
	))
	if err != nil {
		log.Fatalf("Unable to create Sheets client: %v", err)
	}

	// Initialize Google Drive API client
	driveService, err := drive.NewService(ctx, option.WithCredentialsFile(serviceAccountPath), option.WithScopes(
		"https://www.googleapis.com/auth/drive.readonly",
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

	// Fetch and list students from Firestore
	studentDocs, err := iterateStudentsFromDB(firestoreClient, ctx)
	if err != nil {
		log.Fatalf("Error fetching students from Firestore: %v", err)
	}
	log.Printf("Found %d students in Firestore.", len(studentDocs))

	// Get list of student folders from Google Drive
	studentFolders, err := listDriveStudentFolders(driveService, studentsFolderID)
	if err != nil {
		log.Fatalf("Error listing student folders: %v", err)
	}
	if len(studentFolders) == 0 {
		log.Println("No student folders found.")
		return
	}

	// Create a map for quick lookup of folder IDs by name
	folderMap := make(map[string]string)
	for _, folder := range studentFolders {
		folderMap[folder.Name] = folder.Id
	}

	// Iterate over each student document from Firestore
	for _, doc := range studentDocs {
		// Access the student's name from the 'personal' subdocument
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

		log.Printf("Processing student: %s\n", studentName)

		// Check if a corresponding folder exists in Google Drive
		folderID, found := folderMap[studentName]
		if !found {
			log.Printf("No matching folder found in Google Drive for student: %s", studentName)
			continue
		}

		// Get the Answer Sheet ID
		sheetID, err := getAnswerSheetID(driveService, folderID)
		if err != nil {
			log.Printf("Error locating Answer Sheet for %s: %v", studentName, err)
			continue
		}

		// Read data from the student's 'Goals' sheet
		goalsDataRows, err := fetchGoalsSheetData(sheetService, sheetID)
		if err != nil {
			log.Printf("Error reading Goals sheet for %s: %v", studentName, err)
			continue
		}
		if len(goalsDataRows) == 0 {
			log.Printf("No data found in Goals sheet for %s", studentName)
			continue
		}

		// Process and store each row of goals into Firestore
		for _, goalData := range goalsDataRows {
			collegeName, ok := goalData["College"].(string)
			if !ok || collegeName == "" {
				continue
			}

			// Sanitize the college name to create a valid document ID
			sanitizedCollegeName := sanitizeForFirestoreID(collegeName)

			goalsCollection := doc.Ref.Collection("Goals")
			goalDocRef := goalsCollection.Doc(sanitizedCollegeName)

			_, err := goalDocRef.Set(ctx, goalData)
			if err != nil {
				log.Printf("Failed to write goal data for student %s, college %s: %v", studentName, collegeName, err)
				continue
			}
			log.Printf("Successfully wrote goal data for student %s, college %s", studentName, collegeName)
		}
	}

	log.Println("Goals intake process completed.")
}

// Helper function to list student folders from Google Drive
func listDriveStudentFolders(driveService *drive.Service, studentsFolderID string) ([]*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder'", studentsFolderID)
	log.Printf("Querying Google Drive with: %s", query)

	// List files with additional parameters for shared drives
	fileList, err := driveService.Files.List().Q(query).
		Fields("files(id, name)", "nextPageToken").
		SupportsAllDrives(true).         // Include support for shared drives
		IncludeItemsFromAllDrives(true). // Include items from shared drives
		Do()
	if err != nil {
		log.Printf("Error querying Drive API: %v", err) // Detailed error if API fails
		return nil, err
	}

	if len(fileList.Files) == 0 {
		log.Println("No folders found. Double-check that the service account has 'Viewer' access to the Drive folder.")
		log.Println("Ensure that the correct Google Drive folder ID is used and permissions are set correctly.")
	}

	return fileList.Files, nil
}

// Helper function to get the Answer Sheet ID
func getAnswerSheetID(driveService *drive.Service, folderID string) (string, error) {
	// Locate all items within the folder
	query := fmt.Sprintf("'%s' in parents", folderID)
	items, err := driveService.Files.List().Q(query).Fields("files(id, name, mimeType)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Do()
	if err != nil {
		log.Printf("Error querying items in folder ID: %s, %v", folderID, err)
		return "", err
	}

	// Check if a folder named "Test Prep" exists among the items
	var testPrepFolderID string
	for _, item := range items.Files {
		if item.MimeType == "application/vnd.google-apps.folder" && strings.Contains(item.Name, "Test Prep") {
			testPrepFolderID = item.Id
			log.Printf("Found 'Test Prep' folder: %s (ID: %s)", item.Name, item.Id)
			break
		}
	}

	if testPrepFolderID == "" {
		log.Printf("Test Prep folder not found for folder ID: %s", folderID)
		return "", errors.New("Test Prep folder not found")
	}

	// Locate the most recent "Answer Sheet" within the "Test Prep" folder
	query = fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.spreadsheet' and name contains 'Answer Sheet'", testPrepFolderID)
	sheetsList, err := driveService.Files.List().Q(query).OrderBy("createdTime desc").Fields("files(id, name)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Do()
	if err != nil || len(sheetsList.Files) == 0 {
		log.Printf("Answer Sheet not found in Test Prep folder ID: %s", testPrepFolderID)
		return "", errors.New("Answer Sheet not found")
	}

	sheetID := sheetsList.Files[0].Id
	return sheetID, nil
}

// Helper function to read data from a student's 'Goals' sheet in Google Sheets
func fetchGoalsSheetData(sheetService *sheets.Service, sheetID string) ([]map[string]interface{}, error) {
	// Read data from the 'Goals' sheet
	readRange := "Goals!A:G" // Adjusted to read columns A to G
	resp, err := sheetService.Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		log.Printf("Error reading Goals data from sheet ID: %s, Range: %s", sheetID, readRange)
		return nil, err
	}

	// Process each row
	var goalsData []map[string]interface{}
	var startRow int = -1
	dataStarted := false

	// First, determine the starting row (skip headers)
	for i, row := range resp.Values {
		if len(row) == 0 || (len(row) > 0 && strings.TrimSpace(fmt.Sprintf("%v", row[0])) == "") {
			// Empty row, skip
			continue
		}
		firstCell := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", row[0])))
		if firstCell == "school" || firstCell == "college" {
			// Header row detected, skip it
			log.Printf("Skipping header row %d", i+1)
			continue
		} else {
			// Found data row
			startRow = i
			dataStarted = true
			break
		}
	}

	if !dataStarted || startRow == -1 {
		log.Println("No data found in Goals sheet after headers")
		return nil, nil
	}

	// Now process from startRow
	for _, row := range resp.Values[startRow:] {
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

		goalsData = append(goalsData, data)
	}

	return goalsData, nil
}

// Helper function to sanitize a string for use as a Firestore document ID
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

// Helper function to update a subdocument for a student document
func updateSubdocument(firestoreClient *firestore.Client, ctx context.Context, docRef *firestore.DocumentRef, subDocName string, data map[string]interface{}) error {
	_, err := docRef.Set(ctx, map[string]interface{}{subDocName: data}, firestore.MergeAll)
	if err != nil {
		log.Printf("Failed to update subdocument %s: %v", subDocName, err)
		return err
	}
	return nil
}
