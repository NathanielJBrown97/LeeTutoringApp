// backend/internal/dashboard/service.go

package dashboard

import (
	"context"
	"errors"
	"log"
)

// fetchStudentData fetches the student data based on the parent document ID and the selected student ID.
func (a *App) fetchStudentData(parentDocumentID string, selectedStudentID string) (*DashboardData, error) {
	ctx := context.Background()

	// Fetch the parent document using the provided parentDocumentID
	parentDoc, err := a.FirestoreClient.Collection("parents").Doc(parentDocumentID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching parent document: %v", err)
		return nil, err
	}

	// Extract associated_students array from the parent document
	associatedStudents, ok := parentDoc.Data()["associated_students"].([]interface{})
	if !ok || len(associatedStudents) == 0 {
		log.Println("No associated students found or unable to cast associated_students to []interface{}")
		return nil, errors.New("no associated students found")
	}

	var students []Student
	var studentName, teamLead string
	var remainingHours int
	var associatedTutors []string
	var actScores []int64

	for _, s := range associatedStudents {
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
		remainingHoursVal, hoursOk := businessData["remaining_hours"].(int64)
		lead, leadOk := businessData["team_lead"].(string)

		tutors, tutorsOk := businessData["associated_tutors"].([]interface{})
		if tutorsOk {
			associatedTutors = []string{}
			for _, tutor := range tutors {
				if tutorStr, ok := tutor.(string); ok {
					associatedTutors = append(associatedTutors, tutorStr)
				}
			}
		}

		// Fetch the most recent ACT scores
		actDoc, err := studentDoc.Ref.Collection("tests").Doc("most_recent_act").Get(ctx)
		if err == nil {
			if scores, ok := actDoc.Data()["most_recent_act"].([]interface{}); ok {
				actScores = make([]int64, len(scores))
				for i, score := range scores {
					if s, ok := score.(int64); ok {
						actScores[i] = s
					}
				}
			}
		} else {
			log.Printf("Error fetching most recent ACT scores for student ID %s: %v", studentID, err)
		}

		if selectedStudentID == "" || selectedStudentID == studentID {
			if !nameOk || !hoursOk || !leadOk {
				log.Printf("Error: Expected 'name' as string, 'remaining_hours' as int64, and 'team_lead' as string. Got Name OK: %v, Hours OK: %v, Lead OK: %v", nameOk, hoursOk, leadOk)
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
