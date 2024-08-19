package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/NathanielJBrown97/LeeTutoringApp/backend/models"
)

func GetStudent(w http.ResponseWriter, r *http.Request, client *firestore.Client) {
	ctx := context.Background()
	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		http.Error(w, "Missing student_id", http.StatusBadRequest)
		return
	}

	doc, err := client.Collection("students").Doc(studentID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Error(w, "Student not found", http.StatusNotFound)
		} else {
			log.Printf("Failed to get student: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	var student models.Student
	err = doc.DataTo(&student)
	if err != nil {
		log.Printf("Failed to unmarshal student data: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert the student struct to JSON and send it as the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}
