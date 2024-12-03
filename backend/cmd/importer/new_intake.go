package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// StudentData represents the structure for Firestore
type StudentData struct {
	Name              string `json:"name"`
	StudentEmail      string `json:"student_email"`
	StudentNumber     string `json:"student_number"`
	ParentEmail       string `json:"parent_email"`
	ParentNumber      string `json:"parent_number"`
	School            string `json:"school"`
	Grade             string `json:"grade"`
	Scheduler         string `json:"scheduler"`
	TestFocus         string `json:"test_focus"`
	Accommodations    string `json:"accommodations"`
	Interests         string `json:"interests"`
	Availability      string `json:"availability"`
	RegisteredForTest bool   `json:"registered_for_test"`
	TestDate          string `json:"test_date"`
	ClassroomID       string `json:"classroom_id"`
	DriveURL          string `json:"drive_url"` // Only unique ID
}

// FirestoreUpdater handles Firestore operations
type FirestoreUpdater struct {
	Client *firestore.Client
}

func main() {
	// Load environment variables
	err := godotenv.Load()
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

	studentsFolderID := os.Getenv("STUDENTS_FOLDER_ID")
	if studentsFolderID == "" {
		log.Fatal("STUDENTS_FOLDER_ID is not set in the .env file")
	}

	// Initialize Firestore client
	ctx := context.Background()
	firestoreClient, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()

	// Initialize Drive and Sheets services
	driveService, sheetsService, err := initializeGoogleServices(serviceAccountPath)
	if err != nil {
		log.Fatalf("Failed to initialize Google services: %v", err)
	}

	// Test Drive Access within the identified Shared Drive
	testDriveAccess(ctx, driveService, studentsFolderID)

	// Initialize FirestoreUpdater
	fu := FirestoreUpdater{Client: firestoreClient}

	// Iterate through student folders
	err = iterateStudentFolders(ctx, driveService, sheetsService, fu, studentsFolderID)
	if err != nil {
		log.Fatalf("Error during processing: %v", err)
	}

	log.Println("All students have been processed successfully.")
}

// initializeGoogleServices initializes Drive and Sheets services
func initializeGoogleServices(serviceAccountPath string) (*drive.Service, *sheets.Service, error) {
	ctx := context.Background()
	driveService, err := drive.NewService(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create Drive client: %v", err)
	}

	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create Sheets client: %v", err)
	}

	return driveService, sheetsService, nil
}

// testDriveAccess performs a simple access test to verify service account permissions
func testDriveAccess(ctx context.Context, driveService *drive.Service, studentsFolderID string) {
	log.Println("Starting Drive Access Test...")

	folders, err := listFoldersInDrive(ctx, driveService, studentsFolderID)
	if err != nil {
		log.Fatalf("Drive Access Test Failed: %v", err)
	}

	if len(folders) == 0 {
		log.Println("Drive Access Test: No folders found in the Students folder.")
	} else {
		log.Printf("Drive Access Test: Found %d folders in the Students folder.", len(folders))
		for _, folder := range folders {
			log.Printf("Folder - Name: %s, ID: %s", folder.Name, folder.Id)
		}
	}
}

// iterateStudentFolders processes each student folder
func iterateStudentFolders(ctx context.Context, driveService *drive.Service, sheetsService *sheets.Service, fu FirestoreUpdater, studentsFolderID string) error {
	// List all folders in the students folder
	folders, err := listFoldersInDrive(ctx, driveService, studentsFolderID)
	if err != nil {
		return fmt.Errorf("failed to list student folders: %v", err)
	}

	log.Printf("Found %d student folders.", len(folders))

	for _, folder := range folders {
		// **Change 1: Skip 'Previous Students' folder**
		if folder.Name == "Previous Students" {
			log.Println("Skipping 'Previous Students' folder.")
			continue
		}

		log.Printf("Processing folder: %s (ID: %s)", folder.Name, folder.Id)
		err := processStudentFolder(ctx, driveService, sheetsService, fu, folder.Id, folder.Name)
		if err != nil {
			log.Printf("Error processing folder %s: %v", folder.Name, err)
			// Continue with the next folder instead of terminating
			continue
		}
	}

	return nil
}

// listFoldersInDrive lists all folders within a given parent folder in Drive
func listFoldersInDrive(ctx context.Context, driveService *drive.Service, parentID string) ([]*drive.File, error) {
	var folders []*drive.File
	pageToken := ""

	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false", parentID)
	log.Printf("Drive API Query: %s", query)

	for {
		req := driveService.Files.List().
			Q(query).
			Fields("nextPageToken, files(id, name)").
			PageToken(pageToken).
			PageSize(100). // Set a reasonable page size
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true)

		res, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve files: %v", err)
		}

		log.Printf("Retrieved %d folders in current page.", len(res.Files))
		for _, file := range res.Files {
			log.Printf("Folder Found - Name: %s, ID: %s", file.Name, file.Id)
		}

		folders = append(folders, res.Files...)

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	return folders, nil
}

// processStudentFolder processes a single student folder
func processStudentFolder(ctx context.Context, driveService *drive.Service, sheetsService *sheets.Service, fu FirestoreUpdater, studentFolderID, studentFolderName string) error {
	// Find 'Test Prep' subfolder
	testPrepFolder, err := findSubfolder(ctx, driveService, studentFolderID, "Test Prep")
	if err != nil {
		return fmt.Errorf("unable to find 'Test Prep' folder: %v", err)
	}

	if testPrepFolder == nil {
		return fmt.Errorf("'Test Prep' folder not found in student folder %s", studentFolderName)
	}

	// Find the spreadsheet starting with 'Answer Sheet -'
	answerSheet, err := findAnswerSheetSpreadsheet(ctx, driveService, testPrepFolder.Id)
	if err != nil {
		return fmt.Errorf("unable to find 'Answer Sheet -' spreadsheet: %v", err)
	}

	if answerSheet == nil {
		return fmt.Errorf("'Answer Sheet -' spreadsheet not found in 'Test Prep' folder for student %s", studentFolderName)
	}

	// Read data from the spreadsheet (only Profile tab)
	allData, err := readSpreadsheetData(ctx, sheetsService, answerSheet.Id)
	if err != nil {
		return fmt.Errorf("failed to read spreadsheet data: %v", err)
	}

	// **Change 2: Remove 'data' folder access and related data retrieval**

	// **Change 3: Use spreadsheet ID as DriveURL**
	// Prepare StudentData
	studentData, err := prepareStudentData(allData, answerSheet.Id) // Pass spreadsheet ID instead of WebViewLink
	if err != nil {
		return fmt.Errorf("failed to prepare student data: %v", err)
	}

	// Initialize or update student in Firestore
	err = fu.InitializeNewStudent(ctx, studentData)
	if err != nil {
		return fmt.Errorf("failed to initialize student in Firestore: %v", err)
	}

	log.Printf("Successfully initialized student: %s", studentData.Name)
	return nil
}

// findSubfolder finds a subfolder by name within a parent folder
func findSubfolder(ctx context.Context, driveService *drive.Service, parentID, subfolderName string) (*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and name = '%s' and trashed = false", parentID, subfolderName)
	log.Printf("Searching for subfolder '%s' with query: %s", subfolderName, query)
	res, err := driveService.Files.List().
		Q(query).
		Fields("files(id, name)").
		PageSize(10).
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to search for subfolder: %v", err)
	}

	if len(res.Files) == 0 {
		log.Printf("Subfolder '%s' not found in parent ID: %s", subfolderName, parentID)
		return nil, nil // Subfolder not found
	}

	log.Printf("Subfolder '%s' found with ID: %s", subfolderName, res.Files[0].Id)
	return res.Files[0], nil
}

// findAnswerSheetSpreadsheet finds the spreadsheet starting with 'Answer Sheet -' in a folder
func findAnswerSheetSpreadsheet(ctx context.Context, driveService *drive.Service, folderID string) (*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.spreadsheet' and name contains 'Answer Sheet -' and trashed = false", folderID)
	log.Printf("Searching for 'Answer Sheet -' spreadsheet with query: %s", query)
	res, err := driveService.Files.List().
		Q(query).
		Fields("files(id, name, webViewLink)").
		PageSize(10).
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to search for spreadsheets: %v", err)
	}

	if len(res.Files) == 0 {
		log.Println("'Answer Sheet -' spreadsheet not found.")
		return nil, nil // Spreadsheet not found
	}

	// Assuming only one spreadsheet starts with 'Answer Sheet -'
	log.Printf("'Answer Sheet -' spreadsheet found: Name: %s, ID: %s", res.Files[0].Name, res.Files[0].Id)
	return res.Files[0], nil
}

// readSpreadsheetData reads specific cells from the 'Profile' tab in the spreadsheet
func readSpreadsheetData(ctx context.Context, sheetsService *sheets.Service, spreadsheetID string) (map[string]string, error) {
	// Define the range to read. Adjust as needed. Here, B2:B11 corresponds to 10 rows.
	profileRange := "Profile!B2:B11"

	log.Printf("Reading data from spreadsheet ID: %s, Profile Range: %s", spreadsheetID, profileRange)

	// Read Profile data
	profileResp, err := sheetsService.Spreadsheets.Values.Get(spreadsheetID, profileRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from 'Profile' tab: %v", err)
	}

	// Initialize the map to hold all data
	allData := make(map[string]string)

	// Define the keys corresponding to each row in the Profile tab
	profileKeys := []string{
		"student_email",
		"parent_email",
		"name",
		"high_school",
		"grade",
		"test_focus",
		"accommodations",
		"test_date",
		"scheduler",
		"interests",
	}

	// Iterate over each expected key and extract the corresponding value if available
	for i, key := range profileKeys {
		if i < len(profileResp.Values) && len(profileResp.Values[i]) > 0 {
			allData[key] = fmt.Sprintf("%v", profileResp.Values[i][0])
		} else {
			allData[key] = "" // Assign a default empty string if data is missing
			log.Printf("Profile Data - %s: [Missing]", key)
		}
		log.Printf("Profile Data - %s: %s", key, allData[key])
	}

	return allData, nil
}

// prepareStudentData constructs the StudentData struct from profile data and spreadsheet ID
func prepareStudentData(allData map[string]string, spreadsheetID string) (StudentData, error) {
	// Extract and map data
	studentEmail := allData["student_email"]
	parentEmail := allData["parent_email"]
	name := allData["name"]
	highSchool := allData["high_school"]
	grade := allData["grade"]
	testFocus := allData["test_focus"]
	accommodations := allData["accommodations"]
	testDate := allData["test_date"]
	scheduler := allData["scheduler"]
	interests := allData["interests"]
	// classroomID := allData["classroom_id"] // Removed

	// Determine if registered for test
	registeredForTest := false
	if strings.TrimSpace(testDate) != "" {
		registeredForTest = true
	}

	// Use spreadsheet ID as DriveURL (unique identifier)
	uniqueDriveID := spreadsheetID

	// Validate uniqueDriveID (ensure it's not empty)
	if uniqueDriveID == "" {
		log.Printf("Warning: 'DriveURL' (unique ID) is empty for student %s. It will be set as empty in Firestore.", name)
	}

	studentData := StudentData{
		Name:              name,
		StudentEmail:      studentEmail,
		StudentNumber:     "", // Initialize as empty or fetch if available
		ParentEmail:       parentEmail,
		ParentNumber:      "", // Initialize as empty or fetch if available
		School:            highSchool,
		Grade:             grade,
		Scheduler:         scheduler,
		TestFocus:         testFocus,
		Accommodations:    accommodations,
		Interests:         interests,
		Availability:      "", // Initialize as empty or fetch if available
		RegisteredForTest: registeredForTest,
		TestDate:          testDate,
		ClassroomID:       "",            // Set to empty as per requirements
		DriveURL:          uniqueDriveID, // Store only the unique ID
	}

	return studentData, nil
}

// InitializeNewStudent initializes a new student in Firestore
func (fu *FirestoreUpdater) InitializeNewStudent(ctx context.Context, studentData StudentData) error {
	// Generate a unique ID for the new student document
	docRef := fu.Client.Collection("students").NewDoc()
	studentID := docRef.ID

	log.Printf("Generated student ID: %s for student: %s", studentID, studentData.Name)

	// Prepare 'personal' data
	personalData := map[string]interface{}{
		"name":           studentData.Name,
		"student_email":  studentData.StudentEmail,
		"student_number": studentData.StudentNumber,
		"parent_email":   studentData.ParentEmail,
		"parent_number":  studentData.ParentNumber,
		"high_school":    studentData.School,
		"grade":          studentData.Grade,
		"accommodations": studentData.Accommodations,
		"interests":      studentData.Interests,
	}

	// Prepare 'test_appointment' data
	testAppointmentData := map[string]interface{}{
		"registered_for_test": studentData.RegisteredForTest,
		"test_date":           studentData.TestDate,
	}

	// Prepare 'business' data
	businessData := map[string]interface{}{
		"firebase_id":       studentID,
		"scheduler":         studentData.Scheduler,
		"test_focus":        studentData.TestFocus,
		"test_appointment":  testAppointmentData,
		"associated_tutors": []string{}, // Initialize as empty array
		"team_lead":         "",         // Initialize as empty string
		"remaining_hours":   0,          // Initialize as zero
		"lifetime_hours":    0,          // Initialize as zero
		"classroom_id":      studentData.ClassroomID,
		"drive_url":         studentData.DriveURL, // Only the unique ID
	}

	// Combine 'personal' and 'business' into the student document
	studentDocData := map[string]interface{}{
		"personal": personalData,
		"business": businessData,
	}

	// Write the student document to Firestore
	_, err := docRef.Set(ctx, studentDocData)
	if err != nil {
		return fmt.Errorf("failed to save student data: %v", err)
	}

	log.Printf("Successfully initialized student: %s with ID: %s", studentData.Name, studentID)
	return nil
}

// Helper function to safely extract string values from a cell
func getCellStringValue(values [][]interface{}, index int) (string, bool) {
	if index >= len(values) {
		return "", false
	}
	if len(values[index]) == 0 || values[index][0] == nil {
		return "", false
	}

	switch v := values[index][0].(type) {
	case string:
		return strings.TrimSpace(v), true
	case float64:
		// If the cell contains a number but expected a string, convert it to string
		return fmt.Sprintf("%.0f", v), true
	default:
		return "", false
	}
}
