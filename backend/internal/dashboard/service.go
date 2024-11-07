// backend/internal/dashboard/service.go

package dashboard

import (
	"context"
	"errors"
	"log"

	"cloud.google.com/go/firestore"
)

var ErrNoAssociatedStudents = errors.New("no associated students found")

// checkAndAssociateStudents checks if the parent has associated students and performs automatic association if needed
func (a *App) checkAndAssociateStudents(parentDocumentID string, parentEmail string) error {
	ctx := context.Background()

	// Reference to the parent's document
	parentDocRef := a.FirestoreClient.Collection("parents").Doc(parentDocumentID)

	// Fetch the parent document
	parentDocSnap, err := parentDocRef.Get(ctx)
	if err != nil {
		log.Printf("Error fetching parent document: %v", err)
		return err
	}

	// Check if 'associated_students' field exists
	var associatedStudents []interface{}
	if val, exists := parentDocSnap.Data()["associated_students"]; exists {
		associatedStudents, _ = val.([]interface{})
	}

	// If parent already has associated students, do nothing
	if len(associatedStudents) > 0 {
		return nil
	}

	// If no email is available, cannot proceed
	if parentEmail == "" {
		log.Println("Parent email not available in session")
		return ErrNoAssociatedStudents
	}

	// Search for students with matching 'personal.parent_email'
	studentsCollection := a.FirestoreClient.Collection("students")
	query := studentsCollection.Where("personal.parent_email", "==", parentEmail)
	studentDocs, err := query.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error querying students by parent_email: %v", err)
		return err
	}

	// If no students found, return error
	if len(studentDocs) == 0 {
		log.Printf("No students found with parent_email: %s", parentEmail)
		return ErrNoAssociatedStudents
	}

	// Extract student IDs
	var studentIDs []string
	for _, doc := range studentDocs {
		studentIDs = append(studentIDs, doc.Ref.ID)
	}

	// Update the parent's 'associated_students' field
	_, err = parentDocRef.Set(ctx, map[string]interface{}{
		"associated_students": studentIDs,
	}, firestore.MergeAll)
	if err != nil {
		log.Printf("Error updating parent's associated_students: %v", err)
		return err
	}

	log.Printf("Automatically associated students %v with parent %s", studentIDs, parentDocumentID)
	return nil
}

// isStudentAssociatedWithParent checks if the student is associated with the parent
func (a *App) isStudentAssociatedWithParent(ctx context.Context, parentDocumentID string, studentID string) (bool, error) {
	// Fetch the parent document
	parentDoc, err := a.FirestoreClient.Collection("parents").Doc(parentDocumentID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching parent document: %v", err)
		return false, err
	}

	// Extract associated_students array
	associatedStudentsInterface, ok := parentDoc.Data()["associated_students"].([]interface{})
	if !ok || len(associatedStudentsInterface) == 0 {
		log.Println("No associated students found for parent")
		return false, nil
	}

	for _, s := range associatedStudentsInterface {
		sID, ok := s.(string)
		if ok && sID == studentID {
			return true, nil
		}
	}

	return false, nil
}

// fetchStudentData fetches the student data based on the parent document ID and the selected student ID.
func (a *App) fetchStudentData(parentDocumentID string, selectedStudentID string) (*DashboardData, error) {
	ctx := context.Background()

	// Fetch the parent document
	parentDoc, err := a.FirestoreClient.Collection("parents").Doc(parentDocumentID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching parent document: %v", err)
		return nil, err
	}

	// Extract associated_students array
	associatedStudentsInterface, ok := parentDoc.Data()["associated_students"].([]interface{})
	if !ok || len(associatedStudentsInterface) == 0 {
		log.Println("No associated students found for parent")
		return nil, ErrNoAssociatedStudents
	}

	var students []Student
	var studentName, teamLead string
	var remainingHours int
	associatedTutors := []string{}
	actScores := []int64{}

	for _, s := range associatedStudentsInterface {
		studentID, ok := s.(string)
		if !ok {
			log.Println("Error casting student ID to string")
			continue
		}

		// Access the student document in Firestore
		studentDoc, err := a.FirestoreClient.Collection("students").Doc(studentID).Get(ctx)
		if err != nil {
			log.Printf("Error fetching student document with ID %s: %v", studentID, err)
			continue
		}

		personalData, ok := studentDoc.Data()["personal"].(map[string]interface{})
		if !ok {
			log.Printf("Error fetching 'personal' data for student ID %s", studentID)
			continue
		}

		businessData, ok := studentDoc.Data()["business"].(map[string]interface{})
		if !ok {
			log.Printf("Error fetching 'business' data for student ID %s", studentID)
			continue
		}

		name, nameOk := personalData["name"].(string)
		remainingHoursVal, hoursOk := businessData["remaining_hours"].(float64)
		lead, leadOk := businessData["team_lead"].(string)

		tutorsInterface, tutorsOk := businessData["associated_tutors"].([]interface{})
		if tutorsOk {
			associatedTutors = []string{}
			for _, tutor := range tutorsInterface {
				if tutorStr, ok := tutor.(string); ok {
					associatedTutors = append(associatedTutors, tutorStr)
				}
			}
		}

		// Fetch the most recent ACT scores
		// For simplicity, we can fetch recent test scores from 'Test Data' subcollection
		testDataDocs, err := studentDoc.Ref.Collection("Test Data").Documents(ctx).GetAll()
		if err == nil {
			for _, doc := range testDataDocs {
				data := doc.Data()
				if actData, ok := data["ACT"].(map[string]interface{}); ok {
					if totalScore, ok := actData["Total"].(float64); ok {
						actScores = append(actScores, int64(totalScore))
					}
				}
			}
		} else {
			log.Printf("Error fetching Test Data for student ID %s: %v", studentID, err)
		}

		if selectedStudentID == "" || selectedStudentID == studentID {
			if !nameOk || !hoursOk || !leadOk {
				log.Printf("Error: Expected 'name' as string, 'remaining_hours' as float64, and 'team_lead' as string. Got Name OK: %v, Hours OK: %v, Lead OK: %v", nameOk, hoursOk, leadOk)
				return nil, errors.New("unable to retrieve student data")
			}
			studentName = name
			remainingHours = int(remainingHoursVal)
			teamLead = lead
		}

		students = append(students, Student{ID: studentID, Name: name})
	}

	return &DashboardData{
		StudentName:        studentName,
		RemainingHours:     remainingHours,
		TeamLead:           teamLead,
		AssociatedTutors:   associatedTutors,
		AssociatedStudents: students,
		RecentActScores:    actScores,
	}, nil
}
