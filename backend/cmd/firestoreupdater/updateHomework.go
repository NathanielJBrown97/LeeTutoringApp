package firestoreupdater

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/iterator"
)

// HomeworkPayload is the JSON body we expect.
type HomeworkPayload struct {
	StudentName string `json:"studentName"`
	Date        string `json:"date"`
	Percentage  string `json:"percentage"`
	Tutor       string `json:"tutor"`
	Duration    string `json:"duration"`
	Attendance  string `json:"attendance"`
	Feedback    string `json:"feedback"`
}

// UpdateHomeworkCompletionHandler:
//   - Finds the student in Firestore
//   - Opens the subcollection "Homework Completion"
//   - Creates a doc with the date replaced with dashes
//   - Stores exactly these 7 fields: date, percentage, tutor, duration, attendance, feedback, timestamp
func (app *App) UpdateHomeworkCompletionHandler(w http.ResponseWriter, r *http.Request) {
	// Only handle POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the request body
	var payload HomeworkPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received homework completion data: %+v\n", payload)

	ctx := context.Background()
	client := app.FirestoreClient

	// 1) Query 'students' collection to find the doc with the matching name
	iter := client.Collection("students").
		Where("personal.name", "==", payload.StudentName).
		Documents(ctx)

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

	// 2) Access the subcollection "Homework Completion"
	studentDocRef := doc.Ref
	homeworkCollection := studentDocRef.Collection("Homework Completion")

	// 3) Create a doc name from the date by replacing / with -
	docName := strings.ReplaceAll(payload.Date, "/", "-")

	// 4) Exactly these 7 fields (no extra, no less)
	data := map[string]interface{}{
		"date":                payload.Date,
		"percentage_complete": payload.Percentage,
		"tutor":               payload.Tutor,
		"duration":            payload.Duration,
		"attendance":          payload.Attendance,
		"feedback":            payload.Feedback,
		"timestamp":           time.Now().Format(time.RFC3339), // store as string
	}

	_, err = homeworkCollection.Doc(docName).Set(ctx, data)
	if err != nil {
		log.Printf("Failed to write to Firestore: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Homework completion data saved successfully.")
}
