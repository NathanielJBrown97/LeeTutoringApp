package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
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

		// Read data from the student's 'Test Dates' sheet
		testDatesRows, err := fetchTestDatesSheetData(sheetService, sheetID)
		if err != nil {
			log.Printf("Error reading Test Dates sheet for %s: %v", studentName, err)
			continue
		}
		if len(testDatesRows) == 0 {
			log.Printf("No data found in Test Dates sheet for %s", studentName)
			continue
		}

		// Process and store each row of test dates into Firestore
		for _, testData := range testDatesRows {
			testType, ok := testData["Test Type"].(string)
			if !ok || testType == "" {
				continue
			}
			testDateStr, ok := testData["Test Date"].(string)
			if !ok || testDateStr == "" {
				testDateStr = "UnknownDate"
			}

			// Sanitize the date string to avoid slashes in document IDs
			safeTestDateStr := strings.ReplaceAll(testDateStr, "/", "-")
			docName := fmt.Sprintf("%s %s", testType, safeTestDateStr)

			testDatesCollection := doc.Ref.Collection("Test Dates")
			testDateDocRef := testDatesCollection.Doc(docName)

			_, err := testDateDocRef.Set(ctx, testData)
			if err != nil {
				log.Printf("Failed to write test date data for student %s, test %s: %v", studentName, docName, err)
				continue
			}
			log.Printf("Successfully wrote test date data for student %s, test %s", studentName, docName)
		}
	}

	log.Println("Data intake process completed.")
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

// Helper function to read data from a student's 'Test Dates' sheet in Google Sheets
func fetchTestDatesSheetData(sheetService *sheets.Service, sheetID string) ([]map[string]interface{}, error) {
	// Read data from the 'Test Dates' sheet
	readRange := "Test Dates!A2:F" // Assuming headers are in the first row
	resp, err := sheetService.Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		log.Printf("Error reading Test Dates data from sheet ID: %s, Range: %s", sheetID, readRange)
		return nil, err
	}

	// Process each row
	var testDatesData []map[string]interface{}
	for _, row := range resp.Values {
		// Helper function to safely get cell values
		getCellStringValue := func(row []interface{}, index int) string {
			if index >= len(row) {
				return ""
			}
			if row[index] == nil {
				return ""
			}
			return fmt.Sprintf("%v", row[index])
		}

		testType := getCellStringValue(row, 0)
		testDate := getCellStringValue(row, 1)
		regDeadline := getCellStringValue(row, 2)
		lateRegDeadline := getCellStringValue(row, 3)
		scoreReleaseDate := getCellStringValue(row, 4)
		notes := getCellStringValue(row, 5)

		// Skip rows where Test Type or Test Date is missing
		if testType == "" || testDate == "" {
			continue
		}

		data := map[string]interface{}{
			"Test Type":                     testType,
			"Test Date":                     testDate,
			"Regular Registration Deadline": regDeadline,
			"Late Registration Deadline":    lateRegDeadline,
			"Score Release Date":            scoreReleaseDate,
		}

		if notes != "" {
			data["Notes"] = notes
		} else {
			// Optionally, you can initialize an empty 'Notes' field
			data["Notes"] = ""
		}

		testDatesData = append(testDatesData, data)
	}

	return testDatesData, nil
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
