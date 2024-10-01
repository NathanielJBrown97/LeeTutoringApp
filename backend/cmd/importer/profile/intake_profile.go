package main

import (
	"context"
	"errors"
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

func main() {
	// Load environment variables from the .env file
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

		// Read data from the student's Profile sheet
		profileData, err := fetchProfileSheetData(sheetService, driveService, folderID)
		if err != nil {
			log.Printf("Error reading profile sheet for %s: %v", studentName, err)
			continue
		}
		if len(profileData) == 0 {
			log.Printf("No data found in profile sheet for %s", studentName)
			continue
		}

		// Log extracted data
		log.Printf("Extracted profile data for %s: %+v", studentName, profileData)

		// Sanitize extracted data: Treat "-" or empty strings as missing
		sanitize := func(value string) string {
			if value == "-" || value == "" {
				return ""
			}
			return value
		}

		// Prepare updated data for the 'personal' and 'business' subdocuments
		personalUpdate := map[string]interface{}{
			"high_school":    sanitize(profileData["High School"]),
			"grade":          sanitize(profileData["Grade"]),
			"student_email":  sanitize(profileData["Student Email"]),
			"parent_email":   sanitize(profileData["Parent Email"]),
			"accommodations": sanitize(profileData["Accommodations"]),
		}

		businessUpdate := map[string]interface{}{
			"test_focus":       sanitize(profileData["Test Focus"]),
			"registered_tests": sanitize(profileData["Registered Tests"]),
			"remaining_hours":  0,         // Initialize to 0
			"status":           "Unknown", // Set default or adjust based on your needs
		}

		// Ensure existing data such as associated_tutors and team_lead are preserved
		existingData := doc.Data()
		if existingBusiness, ok := existingData["business"].(map[string]interface{}); ok {
			if tutors, ok := existingBusiness["associated_tutors"]; ok {
				businessUpdate["associated_tutors"] = tutors
			} else {
				businessUpdate["associated_tutors"] = []string{} // Set default if not present
			}

			if lead, ok := existingBusiness["team_lead"]; ok {
				businessUpdate["team_lead"] = lead
			} else {
				businessUpdate["team_lead"] = "" // Set default if not present
			}
		}

		// Update Firestore with the new personal and business subdocuments
		if err := updateSubdocument(firestoreClient, ctx, doc.Ref, "personal", personalUpdate); err != nil {
			log.Printf("Failed to update personal data for student %s: %v", studentName, err)
			continue
		}

		if err := updateSubdocument(firestoreClient, ctx, doc.Ref, "business", businessUpdate); err != nil {
			log.Printf("Failed to update business data for student %s: %v", studentName, err)
			continue
		}

		log.Printf("Successfully updated student: %s", studentName)
	}

	log.Println("Profile intake process completed.")
}

// Helper function to iterate over student documents in Firestore
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

// Helper function to list student folders from Google Drive
func listDriveStudentFolders(driveService *drive.Service, studentsFolderID string) ([]*drive.File, error) {
	query := "'" + studentsFolderID + "' in parents and mimeType = 'application/vnd.google-apps.folder'"
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

// Helper function to read data from a student's Profile sheet in Google Sheets
func fetchProfileSheetData(sheetService *sheets.Service, driveService *drive.Service, folderID string) (map[string]string, error) {
	// Locate all items within the folder
	query := "'" + folderID + "' in parents"
	log.Printf("Querying Google Drive for all items in folder ID: %s", folderID)
	items, err := driveService.Files.List().Q(query).Fields("files(id, name, mimeType)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Do()
	if err != nil {
		log.Printf("Error querying items in folder ID: %s, %v", folderID, err)
		return nil, err
	}

	// Log all items found
	log.Printf("Found %d items in folder ID: %s", len(items.Files), folderID)
	for _, item := range items.Files {
		log.Printf("Item found - Name: %s, ID: %s, Type: %s", item.Name, item.Id, item.MimeType)
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
		return nil, errors.New("Test Prep folder not found")
	}

	// Locate the most recent "Answer Sheet" within the "Test Prep" folder
	query = "'" + testPrepFolderID + "' in parents and mimeType = 'application/vnd.google-apps.spreadsheet' and name contains 'Answer Sheet'"
	sheetsList, err := driveService.Files.List().Q(query).OrderBy("createdTime desc").Fields("files(id, name)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Do()
	if err != nil || len(sheetsList.Files) == 0 {
		log.Printf("Answer Sheet not found in Test Prep folder ID: %s", testPrepFolderID)
		return nil, errors.New("Answer Sheet not found")
	}

	// Read from the Profile sheet
	sheetID := sheetsList.Files[0].Id
	readRange := "Profile!A2:B9"
	resp, err := sheetService.Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		log.Printf("Error reading profile data from sheet ID: %s, Range: %s", sheetID, readRange)
		return nil, err
	}

	// Parse the response
	data := make(map[string]string)
	for _, row := range resp.Values {
		if len(row) == 2 {
			key, ok1 := row[0].(string)
			value, ok2 := row[1].(string)
			if ok1 && ok2 {
				key = strings.TrimSpace(key)
				value = strings.TrimSpace(value)
				data[key] = value
				log.Printf("Extracted %s: %s", key, value) // Log each extracted value
			}
		}
	}

	return data, nil
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
