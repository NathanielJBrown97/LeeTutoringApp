// backend/internal/tutordashboard/associate_students.go

package tutordashboard

import (
	"context"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Hardcoded mapping from tutor email to the tutor name as it appears in students' associated_tutors arrays.
var tutorNameMapping = map[string]string{
	"edward@leetutoring.com":    "Edward",
	"eli@leetutoring.com":       "Eli",
	"ben@leetutoring.com":       "Ben",
	"nathaniel@leetutoring.com": "Kyra",
}

// AssociateStudentsForTutor iterates through all student documents in the "students" collection.
// For each student, it looks for the tutor's name (using the mapping) in the business.associated_tutors array.
// If a match is found, it extracts the student's firebase_id (from the "business" subdocument) and
// the student's name (from the "personal" subdocument), then writes a document in the tutor's
// "Associated Students" subcollection. The document ID is set to the student's firebase_id and
// the document stores both the student's name and firebase_id.
// This function now skips writing if the student is already associated.
func AssociateStudentsForTutor(ctx context.Context, client *firestore.Client, tutorUserID string, tutorEmail string) error {
	// Determine the expected tutor name using the mapping.
	tutorName, ok := tutorNameMapping[tutorEmail]
	if !ok {
		log.Printf("No mapping found for tutor email: %s", tutorEmail)
		return nil // Alternatively, you could return an error here.
	}
	log.Printf("Associating students for tutor: %s (%s)", tutorEmail, tutorName)

	// Reference to the tutor document in the "tutors" collection.
	tutorDocRef := client.Collection("tutors").Doc(tutorUserID)
	// Get a reference to the subcollection "Associated Students" under this tutor.
	associatedStudentsColl := tutorDocRef.Collection("Associated Students")

	// Iterate over all student documents in the "students" collection.
	studentsIter := client.Collection("students").Documents(ctx)
	defer studentsIter.Stop()

	for {
		doc, err := studentsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error iterating student docs: %v", err)
			return err
		}

		data := doc.Data()

		// Extract the "business" subdocument.
		business, ok := data["business"].(map[string]interface{})
		if !ok {
			continue
		}

		// Get the student's firebase_id from the business subdocument.
		firebaseID, ok := business["firebase_id"].(string)
		if !ok || firebaseID == "" {
			continue
		}

		// Extract the associated_tutors array from the business subdocument.
		tutorsArr, ok := business["associated_tutors"].([]interface{})
		if !ok {
			continue
		}

		// Check if the expected tutor name exists in the array (case-insensitive, trimmed).
		matched := false
		for _, v := range tutorsArr {
			if nameStr, ok := v.(string); ok {
				if strings.EqualFold(strings.TrimSpace(nameStr), strings.TrimSpace(tutorName)) {
					matched = true
					break
				}
			}
		}
		if matched {
			// Check if this student is already associated.
			studentDocRef := associatedStudentsColl.Doc(firebaseID)
			snap, err := studentDocRef.Get(ctx)
			if err == nil && snap.Exists() {
				log.Printf("Skipping existing associated student with firebase_id %s", firebaseID)
				continue
			} else if err != nil && status.Code(err) != codes.NotFound {
				log.Printf("Error checking associated student for firebase_id %s: %v", firebaseID, err)
				return err
			}

			// Extract the student's name from the "personal" subdocument.
			var studentName string
			if personal, ok := data["personal"].(map[string]interface{}); ok {
				studentName, _ = personal["name"].(string)
			}
			if studentName == "" {
				studentName = "Unknown"
			}

			// Write a document in the "Associated Students" subcollection with document ID = firebaseID.
			writeResult, err := studentDocRef.Set(ctx, map[string]interface{}{
				"name":        studentName,
				"firebase_id": firebaseID,
			})
			if err != nil {
				log.Printf("Failed to set associated student doc for firebase_id %s: %v", firebaseID, err)
				return err
			}
			log.Printf("Associated student with firebase_id %s (%s) for tutor %s. Write time: %v", firebaseID, studentName, tutorEmail, writeResult.UpdateTime)
		}
	}

	log.Printf("Successfully processed student associations for tutor %s", tutorEmail)
	return nil
}

// AssociateStudentsHandler is an HTTP handler that triggers the student association process for a tutor.
// For testing purposes, the tutor's user ID and email are retrieved from query parameters.
// In production, these values should be extracted from the authenticated user's context.
func AssociateStudentsHandler(client *firestore.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tutorUserID := r.URL.Query().Get("tutorUserID")
		tutorEmail := r.URL.Query().Get("tutorEmail")
		if tutorUserID == "" || tutorEmail == "" {
			http.Error(w, "Missing tutorUserID or tutorEmail", http.StatusBadRequest)
			return
		}

		if err := AssociateStudentsForTutor(r.Context(), client, tutorUserID, tutorEmail); err != nil {
			http.Error(w, "Failed to associate students", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Successfully associated students."))
	}
}
