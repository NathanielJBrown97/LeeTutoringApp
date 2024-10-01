// backend/internal/parent/student_intake.go

package parent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentInfo struct {
	StudentID   string
	StudentName string
	CanLink     bool
}

func (a *App) StudentIntakeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the session
	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	// Extract the user ID from the session
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		http.Error(w, "UserID not found in session", http.StatusUnauthorized)
		return
	}

	// Parse the form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	numStudentsStr := r.FormValue("numStudents")
	numStudents, err := strconv.Atoi(numStudentsStr)
	if err != nil {
		http.Error(w, "Invalid number of students", http.StatusBadRequest)
		return
	}

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Prepare to store student names or errors
	var studentInfos []StudentInfo
	var studentNotFound bool

	// Loop through the student IDs and fetch names from Firestore
	for i := 1; i <= numStudents; i++ {
		studentID := r.FormValue(fmt.Sprintf("studentId%d", i))
		if studentID == "" {
			continue
		}

		// Access the student document in Firestore
		docRef := firestoreClient.Collection("students").Doc(studentID)
		doc, err := docRef.Get(context.Background())
		if err != nil {
			if status.Code(err) == codes.NotFound {
				log.Printf("Student ID %s: Document not found", studentID)
				studentInfos = append(studentInfos, StudentInfo{
					StudentID:   studentID,
					StudentName: "Not found",
					CanLink:     false, // Flag that this student cannot be linked
				})
				studentNotFound = true
				continue
			} else {
				log.Printf("Error retrieving student ID %s: %v", studentID, err)
				http.Error(w, fmt.Sprintf("Failed to retrieve student ID %s: %v", studentID, err), http.StatusInternalServerError)
				return
			}
		}

		// Access the personal sub-document data as a map
		personalData, ok := doc.Data()["personal"].(map[string]interface{})
		if !ok {
			log.Printf("Student ID %s: Personal data not found or is not a map", studentID)
			studentInfos = append(studentInfos, StudentInfo{
				StudentID:   studentID,
				StudentName: "No personal data",
				CanLink:     false, // Flag that this student cannot be linked
			})
			continue
		}

		// Extract the "name" field from the personal data
		studentName, ok := personalData["name"].(string)
		if !ok || studentName == "" {
			log.Printf("Student ID %s: Name field not found or is empty", studentID)
			studentInfos = append(studentInfos, StudentInfo{
				StudentID:   studentID,
				StudentName: "No name field",
				CanLink:     false, // Flag that this student cannot be linked
			})
		} else {
			studentInfos = append(studentInfos, StudentInfo{
				StudentID:   studentID,
				StudentName: studentName,
				CanLink:     true, // Flag that this student can be linked
			})
		}
	}

	// Render the student confirmation page
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Confirm Students</title>
</head>
<body>
	<h1>Confirm Students</h1>
	<form action="/confirmLinkStudents" method="POST">
		{{range .StudentInfos}}
			<div>
				{{if .CanLink}}
					<p>Is {{.StudentName}} (ID: {{.StudentID}}) your child?</p>
					<select name="confirm_{{.StudentID}}">
						<option value="yes">Yes</option>
						<option value="no">No</option>
					</select>
				{{else}}
					<p>{{.StudentName}} (ID: {{.StudentID}}) was not found. Please re-enter the correct ID.</p>
				{{end}}
			</div>
		{{end}}
		{{if .StudentNotFound}}
			<p>One or more students were not found. Please correct the IDs and try again.</p>
			<button type="button" onclick="location.href='/parentintake';">Re-enter Student IDs</button>
		{{else}}
			<button type="submit">Link Student</button>
		{{end}}
	</form>
</body>
</html>
`

	// Render the template with the student information
	t, err := template.New("confirm").Parse(tmpl)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, struct {
		StudentInfos    []StudentInfo
		StudentNotFound bool
	}{
		StudentInfos:    studentInfos,
		StudentNotFound: studentNotFound,
	})
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}
