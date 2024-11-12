// backend/internal/dashboard/utils.go

package dashboard

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/NathanielJBrown97/LeeTutoringApp/internal/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// getParentCredentials retrieves credentials from the JWT token
func (a *App) getParentCredentials(r *http.Request) (string, string) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("User is not authenticated.")
		return "", ""
	}

	userID, _ := claims["user_id"].(string)
	email, _ := claims["email"].(string)

	log.Printf("Extracted credentials - userID: %s, email: %s", userID, email)
	return userID, email
}

// getAssociatedStudents retrieves the associated_students array from Firestore
func (a *App) getAssociatedStudents(userID string) ([]string, error) {
	ctx := context.Background()
	parentDocRef := a.FirestoreClient.Collection("parents").Doc(userID)
	parentDocSnap, err := parentDocRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Parent document doesn't exist
			return []string{}, nil
		}
		return nil, err
	}

	associatedStudentsInterface, ok := parentDocSnap.Data()["associated_students"]
	if !ok {
		return []string{}, nil
	}

	var associatedStudents []string
	switch v := associatedStudentsInterface.(type) {
	case []interface{}:
		for _, s := range v {
			if studentID, ok := s.(string); ok {
				associatedStudents = append(associatedStudents, studentID)
			}
		}
	case []string:
		associatedStudents = v
	default:
		log.Printf("Invalid data type for 'associated_students' field: %T", associatedStudentsInterface)
		return nil, fmt.Errorf("invalid data type for 'associated_students' field")
	}

	return associatedStudents, nil
}
