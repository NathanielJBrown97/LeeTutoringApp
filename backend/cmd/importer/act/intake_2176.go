package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
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

	rootFolderID := os.Getenv("STUDENTS_FOLDER_ID")
	if rootFolderID == "" {
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

	// List all subfolders in the root folder
	subFolderIDs, err := listSubFolders(driveService, rootFolderID)
	if err != nil {
		log.Fatalf("Error listing subfolders: %v", err)
	}

	// Iterate over each subfolder
	for _, folderID := range subFolderIDs {
		log.Printf("Processing folder ID: %s", folderID)

		// Find 'Test Prep' folder within the current subfolder
		testPrepFolderID, err := findTestPrepFolder(driveService, folderID)
		if err != nil {
			log.Printf("Error finding Test Prep folder for folder ID %s: %v", folderID, err)
			continue
		}

		// Find the 'Answer Sheet -' within the 'Test Prep' folder
		sheetID, err := findAnswerSheet(driveService, testPrepFolderID)
		if err != nil {
			log.Printf("Error finding Answer Sheet in Test Prep folder ID %s: %v", testPrepFolderID, err)
			continue
		}

		// Fetch ACT scores from the '2176' tab
		scores, err := fetchACTScores(sheetService, sheetID)
		if err != nil {
			log.Printf("Error fetching ACT scores from sheet ID %s: %v", sheetID, err)
			continue
		}

		log.Printf("Successfully fetched ACT scores for folder ID %s: %v", folderID, scores)

		// Retrieve the student reference from Firestore based on folder details
		studentRef, err := getStudentRef(firestoreClient, ctx, folderID, driveService)
		if err != nil {
			log.Printf("Error retrieving student document for folder ID %s: %v", folderID, err)
			continue
		}

		// Update Firestore with ACT scores
		err = updateACTScores(firestoreClient, ctx, studentRef, scores)
		if err != nil {
			log.Printf("Error updating Firestore for folder ID %s: %v", folderID, err)
			continue
		}

		log.Printf("Successfully updated ACT scores in Firestore for folder ID %s.", folderID)
	}

	log.Println("Data intake process completed.")
}

// List all subfolders within a given folder ID
func listSubFolders(driveService *drive.Service, parentFolderID string) ([]string, error) {
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder'", parentFolderID)
	log.Printf("Listing subfolders in folder ID: %s with query: %s", parentFolderID, query)

	folders, err := driveService.Files.List().Q(query).Fields("files(id, name)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Do()
	if err != nil {
		log.Printf("Error listing subfolders: %v", err)
		return nil, err
	}

	var folderIDs []string
	for _, folder := range folders.Files {
		log.Printf("Found subfolder: %s (ID: %s)", folder.Name, folder.Id)
		folderIDs = append(folderIDs, folder.Id)
	}

	return folderIDs, nil
}

// Find the 'Test Prep' folder within the student's folder
func findTestPrepFolder(driveService *drive.Service, folderID string) (string, error) {
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and name = 'Test Prep'", folderID)
	log.Printf("Searching for 'Test Prep' folder in folder ID: %s with query: %s", folderID, query)

	folders, err := driveService.Files.List().Q(query).Fields("files(id, name)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Do()
	if err != nil || len(folders.Files) == 0 {
		log.Printf("Error or no 'Test Prep' folder found: %v", err)
		return "", err
	}

	log.Printf("Found 'Test Prep' folder: %s (ID: %s)", folders.Files[0].Name, folders.Files[0].Id)
	return folders.Files[0].Id, nil
}

// Find the 'Answer Sheet -' Google Sheet within the 'Test Prep' folder
func findAnswerSheet(driveService *drive.Service, testPrepFolderID string) (string, error) {
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.spreadsheet' and name contains 'Answer Sheet -'", testPrepFolderID)
	log.Printf("Searching for 'Answer Sheet -' in Test Prep folder ID: %s with query: %s", testPrepFolderID, query)

	sheets, err := driveService.Files.List().Q(query).OrderBy("createdTime desc").Fields("files(id, name)").SupportsAllDrives(true).IncludeItemsFromAllDrives(true).Do()
	if err != nil || len(sheets.Files) == 0 {
		log.Printf("Error or no 'Answer Sheet -' found: %v", err)
		return "", err
	}

	log.Printf("Found 'Answer Sheet -' sheet: %s (ID: %s)", sheets.Files[0].Name, sheets.Files[0].Id)
	return sheets.Files[0].Id, nil
}

// Retrieve the student reference from Firestore by iterating through each document in the 'students' collection
func getStudentRef(firestoreClient *firestore.Client, ctx context.Context, folderID string, driveService *drive.Service) (*firestore.DocumentRef, error) {
	// Fetch the folder details, ensuring all drives are included
	file, err := driveService.Files.Get(folderID).SupportsAllDrives(true).Fields("name").Do()
	if err != nil {
		log.Printf("Error retrieving folder details for folder ID %s: %v", folderID, err)
		return nil, err
	}

	folderName := strings.TrimSpace(file.Name)
	log.Printf("Processing folder: %s", folderName)

	// Retrieve all documents in the 'students' collection
	iter := firestoreClient.Collection("students").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			log.Printf("Error iterating through students collection: %v", err)
			return nil, err
		}

		// Access the 'personal' sub-document and check the 'name' field
		personal, ok := doc.Data()["personal"].(map[string]interface{})
		if !ok {
			log.Printf("No 'personal' sub-document found for document ID: %s", doc.Ref.ID)
			continue
		}

		studentName, ok := personal["name"].(string)
		if !ok {
			log.Printf("No 'name' field found in 'personal' sub-document for document ID: %s", doc.Ref.ID)
			continue
		}

		// Check if the student name matches the folder name
		if strings.EqualFold(studentName, folderName) {
			log.Printf("Found matching student document for folder name %s: Document ID %s", folderName, doc.Ref.ID)
			return doc.Ref, nil
		}
	}

	log.Printf("No matching student document found for folder ID %s with name %s", folderID, folderName)
	return nil, errors.New("no matching student document found")
}

// Fetch ACT scores from the '2176' tab in the Google Sheet
func fetchACTScores(sheetService *sheets.Service, sheetID string) ([]int64, error) {
	readRange := "2176!F5:F9"
	log.Printf("Fetching ACT scores from sheet ID: %s, Range: %s", sheetID, readRange)

	resp, err := sheetService.Spreadsheets.Values.Get(sheetID, readRange).ValueRenderOption("UNFORMATTED_VALUE").Do()
	if err != nil {
		log.Printf("Error reading ACT data from sheet ID: %s, Range: %s", sheetID, readRange)
		return nil, err
	}

	// Verify that sufficient data exists in the expected range
	if len(resp.Values) < 5 {
		log.Println("Insufficient data in the ACT results tab.")
		return nil, errors.New("insufficient ACT data")
	}

	var scores []int64
	for _, row := range resp.Values {
		if len(row) > 0 {
			switch v := row[0].(type) {
			case float64:
				scores = append(scores, int64(v))
			case string:
				// Try to convert string to a number, if possible
				if num, err := strconv.ParseFloat(v, 64); err == nil {
					scores = append(scores, int64(num))
				} else {
					log.Printf("Failed to parse string score as a number: %s", v)
					scores = append(scores, 0)
				}
			default:
				log.Printf("Unexpected type %T for value: %v", v, v)
				scores = append(scores, 0)
			}
		}
	}

	return scores, nil
}

// Update Firestore with ACT scores in the 'tests' sub-document
func updateACTScores(firestoreClient *firestore.Client, ctx context.Context, studentRef *firestore.DocumentRef, scores []int64) error {
	_, err := studentRef.Collection("tests").Doc("most_recent_act").Set(ctx, map[string]interface{}{
		"most_recent_act": scores,
	}, firestore.MergeAll)
	if err != nil {
		log.Printf("Failed to update ACT scores for student %s: %v", studentRef.ID, err)
	}
	return err
}
