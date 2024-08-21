package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/NathanielJBrown97/LeeTutoringApp/backend/handlers"
	"github.com/NathanielJBrown97/LeeTutoringApp/backend/models"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	ctx := context.Background()

	// Explicitly specify the project ID in the Firebase configuration
	conf := &firebase.Config{
		ProjectID: "lee-tutoring-webapp",
	}

	// Initialize the Firebase app using the service account key file
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}

	// Initialize the Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error initializing Firestore client: %v\n", err)
	}
	defer client.Close()

	// Handle linking students with authentication
	http.HandleFunc("/link-student", func(w http.ResponseWriter, r *http.Request) {
		linkStudentHandler(w, r, ctx, app, client)
	})

	// Handle getting student details
	http.HandleFunc("/get-student", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetStudent(w, r, client)
	})

	// Handle creating a parent document
	http.HandleFunc("/create-parent", func(w http.ResponseWriter, r *http.Request) {
		createParentHandler(w, r, ctx, app, client)
	})

	// Handle getting parent details
	http.HandleFunc("/get-parent", func(w http.ResponseWriter, r *http.Request) {
		getParentHandler(w, r, ctx, client)
	})

	// Serve static files from the frontend directory
	fs := http.FileServer(http.Dir("C:/Users/Obses/OneDrive/Documents/LeeTutoringWork/LeeTutoringApp/frontend/temp"))
	http.Handle("/", fs)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if not set
	}
	log.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func linkStudentHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, app *firebase.App, client *firestore.Client) {
	// Verify ID token
	idToken := r.Header.Get("Authorization")
	if idToken == "" {
		http.Error(w, "Missing ID token", http.StatusUnauthorized)
		return
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		http.Error(w, "Failed to initialize Auth client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		http.Error(w, "Invalid ID token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract user ID and student IDs from the request body
	var requestBody struct {
		StudentIDs []string `json:"student_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var studentNames []string
	for _, studentID := range requestBody.StudentIDs {
		doc, err := client.Collection("students").Doc(studentID).Get(ctx)
		if err != nil {
			http.Error(w, "Student ID not found: "+studentID, http.StatusNotFound)
			return
		}

		studentData := doc.Data()
		studentName, ok := studentData["name"].(string)
		if !ok {
			http.Error(w, "Student name not found for ID: "+studentID, http.StatusNotFound)
			return
		}
		studentNames = append(studentNames, studentName)

		// Link student to the parent account
		_, err = client.Collection("parents").Doc(token.UID).Set(ctx, map[string]interface{}{
			"associated_students_ids": firestore.ArrayUnion(studentID),
		}, firestore.MergeAll)
		if err != nil {
			http.Error(w, "Failed to link student ID: "+studentID, http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success":       true,
		"student_names": studentNames,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createParentHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, app *firebase.App, client *firestore.Client) {
	// Verify ID token
	idToken := r.Header.Get("Authorization")
	if idToken == "" {
		http.Error(w, "Missing ID token", http.StatusUnauthorized)
		return
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		http.Error(w, "Failed to initialize Auth client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		http.Error(w, "Invalid ID token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Decode the request body to get the email and name
	var requestBody struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create or update the parent document in Firestore
	_, err = client.Collection("parents").Doc(token.UID).Set(ctx, map[string]interface{}{
		"email": requestBody.Email,
		"name":  requestBody.Name,
	}, firestore.MergeAll)
	if err != nil {
		http.Error(w, "Failed to create/update parent record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func getParentHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, client *firestore.Client) {
	parentID := r.URL.Query().Get("parent_id")
	if parentID == "" {
		http.Error(w, "Missing parent_id", http.StatusBadRequest)
		return
	}

	doc, err := client.Collection("parents").Doc(parentID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Error(w, "Parent not found", http.StatusNotFound)
		} else {
			log.Printf("Failed to get parent: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	var parent models.Parent
	err = doc.DataTo(&parent)
	if err != nil {
		log.Printf("Failed to unmarshal parent data: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parent)
}
