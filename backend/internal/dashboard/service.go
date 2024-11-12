// backend/internal/dashboard/service.go

package dashboard

import (
	"context"
	"errors"
	"log"
)

var ErrNoAssociatedStudents = errors.New("no associated students found")

// fetchStudentData fetches the student data based on the associated_students and selected student ID.
func (a *App) fetchStudentData(associatedStudents []string, selectedStudentID string) (*DashboardData, error) {
	ctx := context.Background()

	if len(associatedStudents) == 0 {
		log.Println("No associated students found for parent")
		return nil, ErrNoAssociatedStudents
	}

	var students []Student
	var studentName, teamLead string
	var remainingHours int
	var associatedTutors []string
	var actScores []int64

	studentFound := false

	for _, studentID := range associatedStudents {
		log.Printf("Processing student ID: %s", studentID)

		// Access the student document in Firestore
		studentDoc, err := a.FirestoreClient.Collection("students").Doc(studentID).Get(ctx)
		if err != nil {
			log.Printf("Error fetching student document with ID %s: %v", studentID, err)
			continue
		}

		studentData := studentDoc.Data()

		// Fetch 'personal' data
		personalData, ok := studentData["personal"].(map[string]interface{})
		if !ok {
			log.Printf("Error fetching 'personal' data for student ID %s", studentID)
			personalData = make(map[string]interface{})
		}

		// Fetch 'business' data
		businessData, ok := studentData["business"].(map[string]interface{})
		if !ok {
			log.Printf("Error fetching 'business' data for student ID %s", studentID)
			businessData = make(map[string]interface{})
		}

		// Extract name
		name, nameOk := personalData["name"].(string)
		if !nameOk {
			log.Printf("Name not found or not a string for student ID %s", studentID)
			name = "Unknown Student"
		}

		// Extract remaining_hours
		var remainingHoursInt int
		if remainingHoursVal, ok := businessData["remaining_hours"]; ok {
			switch v := remainingHoursVal.(type) {
			case float64:
				remainingHoursInt = int(v)
			case int64:
				remainingHoursInt = int(v)
			case int:
				remainingHoursInt = v
			default:
				log.Printf("Unexpected type for remaining_hours: %T", remainingHoursVal)
			}
		} else {
			log.Printf("remaining_hours not found for student ID %s", studentID)
		}

		// Extract team_lead
		lead, leadOk := businessData["team_lead"].(string)
		if !leadOk {
			log.Printf("team_lead not found or not a string for student ID %s", studentID)
			lead = "Unknown Team Lead"
		}

		// Extract associated_tutors
		var studentTutors []string
		if tutorsInterface, tutorsOk := businessData["associated_tutors"]; tutorsOk {
			switch v := tutorsInterface.(type) {
			case []interface{}:
				for _, tutor := range v {
					if tutorStr, ok := tutor.(string); ok {
						studentTutors = append(studentTutors, tutorStr)
					}
				}
			case []string:
				studentTutors = v
			default:
				log.Printf("Unexpected type for associated_tutors: %T", tutorsInterface)
			}
		}

		// Fetch the most recent ACT scores
		testDataDocs, err := studentDoc.Ref.Collection("Test Data").Documents(ctx).GetAll()
		if err == nil {
			for _, doc := range testDataDocs {
				data := doc.Data()
				if actData, ok := data["ACT"].(map[string]interface{}); ok {
					if totalScoreVal, ok := actData["Total"]; ok {
						switch v := totalScoreVal.(type) {
						case float64:
							actScores = append(actScores, int64(v))
						case int64:
							actScores = append(actScores, v)
						case int:
							actScores = append(actScores, int64(v))
						default:
							log.Printf("Unexpected type for ACT Total score: %T", totalScoreVal)
						}
					}
				}
			}
		} else {
			log.Printf("Error fetching Test Data for student ID %s: %v", studentID, err)
		}

		students = append(students, Student{ID: studentID, Name: name})

		if selectedStudentID == "" || selectedStudentID == studentID {
			studentName = name
			remainingHours = remainingHoursInt
			teamLead = lead
			associatedTutors = studentTutors
			studentFound = true
		}
	}

	if !studentFound {
		return nil, errors.New("selected student not found")
	}

	return &DashboardData{
		StudentName:        studentName,
		RemainingHours:     remainingHours,
		TeamLead:           teamLead,
		AssociatedTutors:   associatedTutors,
		AssociatedStudents: students,
		RecentActScores:    actScores,
		NeedsStudentIntake: false,
	}, nil
}
