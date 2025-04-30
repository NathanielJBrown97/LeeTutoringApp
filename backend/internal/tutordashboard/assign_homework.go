package tutordashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	classroom "google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

// HomeworkRequest represents the JSON payload sent from the frontend.
type HomeworkRequest struct {
	Test            string `json:"test"`
	Section         string `json:"section"`
	Topic           string `json:"topic"` // Used for quiz assignments
	Date            string `json:"date"`  // Expected format: "YYYY-MM-DD"
	Timed           bool   `json:"timed"`
	Notes           bool   `json:"notes"`
	Form            string `json:"form"` // Identifier for the form (e.g., test id)
	Work            string `json:"work"` // Problems, passages, or instructions
	ClassID         string `json:"class_id"`
	StudentFolderID string `json:"student_folder_id"`
}

// DueDate holds parsed date information.
type DueDate struct {
	Year  int
	Month int
	Day   int
}

// AssignHomeworkHandler handles incoming requests to assign homework.
func AssignHomeworkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HomeworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	dueDate, err := parseDate(req.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Initialize the Classroom API client.
	ctx := context.Background()
	// Assumes that credentials are provided via the GOOGLE_APPLICATION_CREDENTIALS env variable.
	svc, err := classroom.NewService(ctx, option.WithScopes(
		classroom.ClassroomCourseworkStudentsScope,
		classroom.ClassroomCourseworkMeScope,
	))
	if err != nil {
		log.Printf("Failed to create classroom service: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Construct the assignment title and description.
	title := fmt.Sprintf("%s Homework Due %d.%d", capitalizeFirstLetter(req.Section), dueDate.Month, dueDate.Day)
	descriptionText := buildDescription(req)

	// Build the coursework object.
	// For materials, we're using a SharedDriveFile.
	courseWork := &classroom.CourseWork{
		Title:       title,
		Description: descriptionText,
		Materials: []*classroom.Material{
			{
				DriveFile: &classroom.SharedDriveFile{
					DriveFile: &classroom.DriveFile{
						Id: req.StudentFolderID,
					},
					ShareMode: "VIEW",
				},
			},
		},

		DueDate: &classroom.Date{
			Year:  int64(dueDate.Year),
			Month: int64(dueDate.Month),
			Day:   int64(dueDate.Day),
		},
		DueTime: &classroom.TimeOfDay{
			Hours:   23,
			Minutes: 59,
			Seconds: 59,
		},
		MaxPoints: 100,
		WorkType:  "ASSIGNMENT",
	}

	createdCourseWork, err := svc.Courses.CourseWork.Create(req.ClassID, courseWork).Do()
	if err != nil {
		log.Printf("Failed to create coursework: %v", err)
		http.Error(w, "Failed to create assignment", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":     "success",
		"courseWork": createdCourseWork,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseDate converts a "YYYY-MM-DD" string into a DueDate struct.
func parseDate(dateStr string) (DueDate, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return DueDate{}, err
	}
	return DueDate{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
	}, nil
}

// capitalizeFirstLetter returns the string with the first letter capitalized.
func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	// Assumes ASCII letters.
	return string(s[0]-32) + s[1:]
}

// buildDescription constructs the assignment description based on the request.
func buildDescription(req HomeworkRequest) string {
	description := "Please print the attached PDF and follow the instructions below:\n\n"

	if req.Timed {
		description += "This assignment is timed. Please adhere to the allocated time limits.\n\n"
	}

	if req.Notes {
		description += "Make sure to take notes and mark any problems you're unsure about.\n\n"
	}

	description += fmt.Sprintf("Complete the following problems/passages: %s.\n\n", req.Work)
	description += "Enter your answers into the provided form and submit before the due date."

	return description
}
