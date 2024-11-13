// backend/internal/parent/automatic_association.go

package parent

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/middleware"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) AttemptAutomaticAssociation(w http.ResponseWriter, r *http.Request) {
	// Authentication is handled via middleware
	userID, err := middleware.ExtractUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized: User ID not found", http.StatusUnauthorized)
		return
	}

	// Get user's email from JWT claims
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: User claims not found", http.StatusUnauthorized)
		return
	}
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		http.Error(w, "Email not found in token", http.StatusUnauthorized)
		return
	}

	// Fetch parent document from Firestore
	parentDocRef := a.FirestoreClient.Collection("parents").Doc(userID)
	_, err = parentDocRef.Get(r.Context())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Parent document doesn't exist, create it
			_, err := parentDocRef.Set(context.Background(), map[string]interface{}{
				"associated_students": []string{},
			})
			if err != nil {
				http.Error(w, "Failed to create parent document", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to retrieve parent data", http.StatusInternalServerError)
			return
		}
	}

	// Attempt automatic association
	// Query students where personal.parent_email == email
	studentsQuery := a.FirestoreClient.Collection("students").Where("personal.parent_email", "==", email)
	studentsIter := studentsQuery.Documents(r.Context())
	defer studentsIter.Stop()

	var foundStudentIDs []interface{}
	for {
		doc, err := studentsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, "Error querying students", http.StatusInternalServerError)
			return
		}
		foundStudentIDs = append(foundStudentIDs, doc.Ref.ID)
	}

	if len(foundStudentIDs) > 0 {
		// Update parent's associated_students field
		_, err := parentDocRef.Set(context.Background(), map[string]interface{}{
			"associated_students": firestore.ArrayUnion(foundStudentIDs...),
		}, firestore.MergeAll)
		if err != nil {
			http.Error(w, "Failed to update parent document", http.StatusInternalServerError)
			return
		}
	}

	// Return the associated students
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"associatedStudents": foundStudentIDs,
	})
}
