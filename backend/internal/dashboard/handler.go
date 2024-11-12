// backend/internal/dashboard/handler.go

package dashboard

import (
	"encoding/json"
	"log"
	"net/http"
)

// DashboardData represents the data structure for rendering the dashboard
type DashboardData struct {
	StudentName        string    `json:"studentName"`
	RemainingHours     int       `json:"remainingHours"`
	TeamLead           string    `json:"teamLead"`
	AssociatedTutors   []string  `json:"associatedTutors"`
	AssociatedStudents []Student `json:"associatedStudents"`
	RecentActScores    []int64   `json:"recentActScores"`
	NeedsStudentIntake bool      `json:"needsStudentIntake"`
}

// Student represents a student's basic details.
type Student struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Handler serves the dashboard data as JSON
func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	// Retrieve credentials from the JWT token
	parentDocumentID, parentEmail := a.getParentCredentials(r)
	if parentDocumentID == "" || parentEmail == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
		return
	}

	// Fetch associated_students from Firestore
	associatedStudents, err := a.getAssociatedStudents(parentDocumentID)
	if err != nil {
		log.Printf("Error fetching associated students: %v", err)
		http.Error(w, "Unable to fetch associated students", http.StatusInternalServerError)
		return
	}

	if len(associatedStudents) == 0 {
		// Return a JSON response indicating that student intake is needed
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(DashboardData{
			NeedsStudentIntake: true,
		})
		if err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			http.Error(w, "Unable to encode JSON response", http.StatusInternalServerError)
		}
		return
	}

	// Get selected student ID from query parameters
	selectedStudentID := r.URL.Query().Get("student_id")

	// Fetch the student data using associatedStudents and selectedStudentID
	data, err := a.fetchStudentData(associatedStudents, selectedStudentID)
	if err != nil {
		log.Printf("Error fetching student data: %v", err)
		http.Error(w, "Unable to fetch student data", http.StatusInternalServerError)
		return
	}

	// Set response headers and encode data as JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Unable to encode JSON response", http.StatusInternalServerError)
		return
	}
}
