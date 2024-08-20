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

func GetParent(w http.ResponseWriter, r *http.Request, client *firestore.Client) {
	ctx := context.Background()
	parentID := r.URL.Query().Get("parent_id")
	if parentID == "" {
		http.Error(w, "Missing parent_id", http.StatusBadRequest)
		return
	}

	// Get the document from the "parents" collection
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

	// Map the document data to the Parent struct
	var parent models.Parent
	err = doc.DataTo(&parent)
	if err != nil {
		log.Printf("Failed to unmarshal parent data: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Respond with JSON data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parent)
}
