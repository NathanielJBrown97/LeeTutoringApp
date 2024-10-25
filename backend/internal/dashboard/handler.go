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
	// Retrieve the parent document ID and email from the session
	parentDocumentID, parentEmail := a.getParentCredentials(r)
	if parentDocumentID == "" || parentEmail == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
		return
	}

	// Check and perform automatic association if needed
	err := a.checkAndAssociateStudents(parentDocumentID, parentEmail)
	if err != nil {
		if err == ErrNoAssociatedStudents {
			// Return a JSON response indicating that student intake is needed
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(DashboardData{
				NeedsStudentIntake: true,
			})
			if err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				http.Error(w, "Unable to encode JSON response", http.StatusInternalServerError)
			}
			return
		}
		log.Printf("Error during automatic student association: %v", err)
		http.Error(w, "Unable to associate students", http.StatusInternalServerError)
		return
	}

	// Get selected student ID from query parameters
	selectedStudentID := r.URL.Query().Get("student_id")

	// Fetch the student data using the parentDocumentID and selectedStudentID
	data, err := a.fetchStudentData(parentDocumentID, selectedStudentID)
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

// getParentCredentials retrieves the parent document ID and email based on the logged-in user session
func (a *App) getParentCredentials(r *http.Request) (string, string) {
	// Retrieve the session using the session store
	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		log.Printf("Failed to retrieve session: %v", err)
		return "", ""
	}
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		log.Println("User ID not found in session")
		return "", ""
	}
	userEmail, ok := session.Values["user_email"].(string)
	if !ok || userEmail == "" {
		log.Println("User email not found in session")
		return userID, ""
	}
	return userID, userEmail
}
