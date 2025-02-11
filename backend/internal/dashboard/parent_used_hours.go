package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"
)

// UpdateParentUsedHoursResponse is the JSON response body
type UpdateParentUsedHoursResponse struct {
	ParentID        string  `json:"parent_id"`
	ParentRemaining float64 `json:"parent_remaining_hours"`
	Message         string  `json:"message"`
}

// UpdateParentUsedHoursHandler recalculates all students' lifetime_hours,
// sums them for a parent, then updates the parent's remaining_hours and
// each student's remaining_hours accordingly.
func (a *App) UpdateParentUsedHoursHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// 1. Retrieve parent’s user ID from JWT
	parentID, _ := a.getParentCredentials(r)
	if parentID == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
		return
	}

	// 2. Fetch parent doc from 'parents' collection
	parentDocRef := a.FirestoreClient.Collection("parents").Doc(parentID)
	parentSnap, err := parentDocRef.Get(ctx)
	if err != nil || !parentSnap.Exists() {
		log.Printf("Error or missing parent doc for user %s: %v", parentID, err)
		http.Error(w, "Parent document not found", http.StatusNotFound)
		return
	}

	parentData := parentSnap.Data()

	// -- Extract associated_students from parent data
	associatedStudents, ok := parentData["associated_students"].([]interface{})
	if !ok {
		http.Error(w, "No associated_students found for parent", http.StatusBadRequest)
		return
	}

	// -- Extract qboCustomerId from parent’s business sub-doc
	businessData, ok := parentData["business"].(map[string]interface{})
	if !ok {
		http.Error(w, "No business sub-doc found for parent", http.StatusBadRequest)
		return
	}

	qboCustomerID, ok := businessData["qboCustomerId"].(string)
	if !ok || qboCustomerID == "" {
		http.Error(w, "Invalid or missing qboCustomerId in parent's business sub-doc", http.StatusBadRequest)
		return
	}

	// 3. Fetch parent's total_hours from 'intuit' collection
	intuitDocRef := a.FirestoreClient.Collection("intuit").Doc(qboCustomerID)
	intuitSnap, err := intuitDocRef.Get(ctx)
	if err != nil || !intuitSnap.Exists() {
		log.Printf("Error or missing intuit doc for qboCustomerId %s: %v", qboCustomerID, err)
		http.Error(w, "Intuit data not found", http.StatusNotFound)
		return
	}
	intuitData := intuitSnap.Data()

	// This is the parent's total hours "pool"
	parentTotalHours, _ := intuitData["total_hours"].(float64)

	// 4. For each student, force-update their lifetime_hours (like UpdateUsedHoursHandler),
	//    then sum it up to get parent's total used hours.
	var parentUsedHours float64

	for _, s := range associatedStudents {
		studentID, ok := s.(string)
		if !ok || studentID == "" {
			continue
		}

		// We replicate the logic from UpdateUsedHoursHandler *in code*:
		// a) Sum durations from "Homework Completion"
		// b) Update the student's doc => business.lifetime_hours
		studentLifetime, err := a.forceUpdateStudentUsedHours(ctx, studentID)
		if err != nil {
			log.Printf("Skipping student %s due to error: %v", studentID, err)
			continue
		}

		parentUsedHours += studentLifetime
	}

	// 5. parent's remaining_hours = parentTotalHours - parentUsedHours
	parentRemaining := parentTotalHours - parentUsedHours
	if parentRemaining < 0 {
		parentRemaining = 0 // or allow negative to indicate overage, your choice
	}

	// 6. Update parent's doc => business.remaining_hours
	_, err = parentDocRef.Update(ctx, []firestore.Update{
		{
			Path:  "business.remaining_hours",
			Value: parentRemaining,
		},
	})
	if err != nil {
		log.Printf("Error updating parent's remaining_hours for %s: %v", parentID, err)
		http.Error(w, "Failed to update parent's remaining_hours", http.StatusInternalServerError)
		return
	}

	// 7. Also update each student’s doc => business.remaining_hours
	for _, s := range associatedStudents {
		studentID, ok := s.(string)
		if !ok || studentID == "" {
			continue
		}
		studentDocRef := a.FirestoreClient.Collection("students").Doc(studentID)
		_, err = studentDocRef.Update(ctx, []firestore.Update{
			{
				Path:  "business.remaining_hours",
				Value: parentRemaining,
			},
		})
		if err != nil {
			log.Printf("Error updating student's remaining_hours for %s: %v", studentID, err)
		}
	}

	// 8. Return a simple JSON response
	w.Header().Set("Content-Type", "application/json")
	resp := UpdateParentUsedHoursResponse{
		ParentID:        parentID,
		ParentRemaining: parentRemaining,
		Message:         "Parent used-hours, Student used-hours, and Remaining Hours updated successfully.",
	}
	json.NewEncoder(w).Encode(resp)
}

// forceUpdateStudentUsedHours is an internal helper method
// that replicates the logic from your "UpdateUsedHoursHandler"
// to recalculate a single student's 'lifetime_hours' by summing
// durations in 'Homework Completion' subcollection.
func (a *App) forceUpdateStudentUsedHours(ctx context.Context, studentID string) (float64, error) {
	// 1. Grab the student doc reference
	studentDocRef := a.FirestoreClient.Collection("students").Doc(studentID)

	// 2. Fetch "Homework Completion" subcollection
	hwDocs, err := studentDocRef.Collection("Homework Completion").Documents(ctx).GetAll()
	if err != nil {
		return 0, err
	}

	// 3. Sum up durations (stored as string)
	var totalHours float64
	for _, doc := range hwDocs {
		data := doc.Data()
		durationVal, found := data["duration"]
		if !found {
			continue
		}
		durationStr, ok := durationVal.(string)
		if !ok {
			log.Printf("Skipping doc %s in 'Homework Completion' for student %s: 'duration' not a string", doc.Ref.ID, studentID)
			continue
		}
		parsed, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			log.Printf("Skipping doc %s for student %s: unable to parse 'duration'='%s'", doc.Ref.ID, studentID, durationStr)
			continue
		}
		totalHours += parsed
	}

	// 4. Update student's doc => business.lifetime_hours
	_, err = studentDocRef.Update(ctx, []firestore.Update{
		{
			Path:  "business.lifetime_hours",
			Value: totalHours,
		},
	})
	if err != nil {
		return 0, err
	}

	// 5. Return the recalculated total
	return totalHours, nil
}
