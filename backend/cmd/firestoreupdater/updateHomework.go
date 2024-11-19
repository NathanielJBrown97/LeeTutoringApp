package firestoreupdater

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"google.golang.org/api/iterator"
)

type HomeworkCompletion struct {
	StudentName string `json:"studentName"`
	Date        string `json:"date"`
	Percentage  string `json:"percentage"`
}

func (app *App) UpdateHomeworkCompletionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body
	var hwComp HomeworkCompletion
	err := json.NewDecoder(r.Body).Decode(&hwComp)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client := app.FirestoreClient

	// Query the 'students' collection for the student with matching name
	iter := client.Collection("students").Where("personal.name", "==", hwComp.StudentName).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query Firestore: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	studentDocRef := doc.Ref

	// Access the 'Homework Completion' subcollection
	homeworkCompCollection := studentDocRef.Collection("Homework Completion")

	// Create or update the document with the date as the document ID
	docRef := homeworkCompCollection.Doc(hwComp.Date)

	data := map[string]interface{}{
		"date":       hwComp.Date,
		"percentage": hwComp.Percentage,
	}

	_, err = docRef.Set(ctx, data)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}
