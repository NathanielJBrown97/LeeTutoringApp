package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
)

func (a *App) UpdateStudentLifetimeHoursHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := a.getParentCredentials(r)
	if userID == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
		return
	}

	associatedStudents, err := a.getAssociatedStudents(userID)
	if err != nil {
		log.Printf("Error fetching associated students: %v", err)
		http.Error(w, "Unable to fetch associated students", http.StatusInternalServerError)
		return
	}
	if len(associatedStudents) == 0 {
		http.Error(w, "No associated students found", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	studentID := vars["student_id"]
	if studentID == "" {
		http.Error(w, "Student ID is required", http.StatusBadRequest)
		return
	}

	isAssociated := false
	for _, sID := range associatedStudents {
		if sID == studentID {
			isAssociated = true
			break
		}
	}
	if !isAssociated {
		http.Error(w, "Unauthorized access to student data", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	studentDocRef := a.FirestoreClient.Collection("students").Doc(studentID)
	homeworkCompletionDocs, err := studentDocRef.Collection("Homework Completion").Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching 'Homework Completion' subcollection: %v", err)
		http.Error(w, "Error fetching homework completion data", http.StatusInternalServerError)
		return
	}
	var totalHours float64
	for _, doc := range homeworkCompletionDocs {
		data := doc.Data()
		durationVal, ok := data["duration"]
		if !ok {
			continue
		}

		durationStr, ok := durationVal.(string)
		if !ok {
			log.Printf("Warning: 'duration' for doc %s was not a string", doc.Ref.ID)
			continue
		}

		parsed, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			log.Printf("Warning: could not parse 'duration'='%s' for doc %s: %v", durationStr, doc.Ref.ID, err)
			continue
		}
		totalHours += parsed
	}

	_, err = studentDocRef.Update(ctx, []firestore.Update{
		{
			Path:  "business.lifetime_hours",
			Value: totalHours,
		},
	})
	if err != nil {
		log.Printf("Error updating lifetime_hours for student %s: %v", studentID, err)
		http.Error(w, "Failed to update lifetime_hours", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"student_id":     studentID,
		"lifetime_hours": totalHours,
		"message":        "Lifetime hours updated successfully",
	})
}
