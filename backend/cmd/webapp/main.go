package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore" // Import Firestore client
	firebase "firebase.google.com/go"
	"github.com/NathanielJBrown97/LeeTutoringApp/handlers"
	"google.golang.org/api/option"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Initialize the Firebase app using the service account key file
	opt := option.WithCredentialsFile("backend/serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v\n", err)
	}

	// Initialize the Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error initializing Firestore client: %v\n", err)
	}
	defer client.Close() // Close the client when the main function finishes

	// Handle routes
	http.HandleFunc("/link-student", func(w http.ResponseWriter, r *http.Request) {
		// Example logic to link STUDENTID
		// Here you would extract the necessary data from the request, such as the user ID and the student ID
		// Then, you could write that data to Firestore

		// Example: assuming you send user ID and student ID in the request body (as JSON)
		// Extract user ID and student ID from the request (this is a simplified example)
		studentID := r.URL.Query().Get("student_id")
		userID := r.URL.Query().Get("user_id")

		if studentID == "" || userID == "" {
			http.Error(w, "Missing student_id or user_id", http.StatusBadRequest)
			return
		}

		// Firestore document path: "students/{studentID}"
		_, err := client.Collection("students").Doc(studentID).Set(ctx, map[string]interface{}{
			"userID": userID,
		}, firestore.MergeAll)

		if err != nil {
			http.Error(w, "Failed to link student ID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Successfully linked student ID %s with user ID %s", studentID, userID)
	})

	http.HandleFunc("/get-student", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetStudent(w, r, client)
	})

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
