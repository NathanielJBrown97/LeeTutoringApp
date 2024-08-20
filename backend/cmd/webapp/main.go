package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore" // Import Firestore client
	firebase "firebase.google.com/go"
	"github.com/NathanielJBrown97/LeeTutoringApp/backend/handlers"
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

	// HANDLE - LINKING STUDENTS WITH AUTHENTICATION
	http.HandleFunc("/link-student", func(w http.ResponseWriter, r *http.Request) {
		// Verify ID token
		idToken := r.Header.Get("Authorization")
		if idToken == "" {
			http.Error(w, "Missing ID token", http.StatusUnauthorized)
			return
		}

		authClient, err := app.Auth(ctx)
		if err != nil {
			http.Error(w, "Failed to initialize Auth client: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to initialize Auth client: %v\n", err)
			return
		}

		token, err := authClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			http.Error(w, "Invalid ID token: "+err.Error(), http.StatusUnauthorized)
			log.Printf("Invalid ID token: %v\n", err)
			return
		}

		fmt.Printf("Authenticated user ID: %s\n", token.UID)

		// Extract user ID and student IDs from the request (e.g., from JSON body)
		type RequestBody struct {
			UserID     string   `json:"user_id"`
			StudentIDs []string `json:"student_ids"`
		}

		var requestBody RequestBody
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			log.Printf("Invalid request body: %v\n", err)
			return
		}

		// Link each student ID to the user
		for _, studentID := range requestBody.StudentIDs {
			_, err := client.Collection("students").Doc(studentID).Set(ctx, map[string]interface{}{
				"userID": requestBody.UserID,
			}, firestore.MergeAll)

			if err != nil {
				http.Error(w, "Failed to link student ID "+studentID+": "+err.Error(), http.StatusInternalServerError)
				log.Printf("Failed to link student ID %s: %v\n", studentID, err)
				return
			}
		}

		fmt.Fprintf(w, "Successfully linked student IDs %v with user ID %s", requestBody.StudentIDs, requestBody.UserID)
	})

	// HANDLE - GET STUDENT
	http.HandleFunc("/get-student", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetStudent(w, r, client)
	})

	// Start server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
