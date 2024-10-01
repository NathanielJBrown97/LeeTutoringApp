// backend/internal/dashboard/handler.go

package dashboard

import (
	"html/template"
	"log"
	"net/http"
)

// DashboardData represents the data structure for rendering the dashboard
type DashboardData struct {
	StudentName        string
	RemainingHours     int
	TeamLead           string
	AssociatedTutors   []string
	AssociatedStudents []Student
	RecentActScores    []int64
}

// Student represents a student's basic details.
type Student struct {
	ID   string
	Name string
}

// Handler renders the parent dashboard
func (a *App) Handler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the parent document ID from the session
	parentDocumentID := a.getParentDocumentID(r)
	if parentDocumentID == "" {
		http.Error(w, "Unable to identify parent user", http.StatusUnauthorized)
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

	// Render the template with the retrieved data
	err = a.RenderTemplate(w, data)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
		return
	}
}

// getParentDocumentID retrieves the parent document ID based on the logged-in user session
func (a *App) getParentDocumentID(r *http.Request) string {
	// Retrieve the session using the session store
	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		log.Printf("Failed to retrieve session: %v", err)
		return ""
	}
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		log.Println("User ID not found in session")
		return ""
	}
	return userID
}

// RenderTemplate renders the dashboard template with the provided data
func (a *App) RenderTemplate(w http.ResponseWriter, data *DashboardData) error {
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Parent Dashboard</title>
	</head>
	<body>
		<h1>Welcome to the Parent Dashboard</h1>
		<p>Student Name: {{.StudentName}}</p>
		<p>Remaining Hours: {{.RemainingHours}}</p>
		<p>Team Lead: {{.TeamLead}}</p>
		<h2>Associated Tutors</h2>
		<ul>
			{{range .AssociatedTutors}}
				<li>{{.}}</li>
			{{end}}
		</ul>
		<h2>Associated Students</h2>
		<ul>
			{{range .AssociatedStudents}}
				<li>{{.Name}}</li>
			{{end}}
		</ul>
		<h2>Recent ACT Scores</h2>
		<ul>
			{{range .RecentActScores}}
				<li>{{.}}</li>
			{{end}}
		</ul>
	</body>
	</html>
	`

	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		return err
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return err
	}

	return nil
}
