package tutordashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/iterator"
)

// FetchStudentsByNamesRequest is the expected structure of the request body,
// containing a list of full names to match against `personal.name`.
type FetchStudentsByNamesRequest struct {
	Names []string `json:"names"`
}

// FetchStudentsByNamesHandler handles POST /api/tutor/fetch-students-by-names.
// It reads a JSON body containing an array of full names. Then, it iterates
// through all students in Firestore, checking if `personal.name` matches
// any of the requested names. For each match, it fetches all subcollections
// (Homework Completion, Test Data, etc.) and returns the aggregated data.
func (a *App) FetchStudentsByNamesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse the request body
	var req FetchStudentsByNamesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	// Build a quick lookup set of the requested names
	requestedNames := make(map[string]bool)
	for _, nm := range req.Names {
		requestedNames[strings.TrimSpace(nm)] = true
	}

	// 2. Optionally, you might want to validate the tutor's identity here,
	//    or ensure they are authorized to see these students. If you rely
	//    on the tutor's Firestore user doc, you'd parse from token or query
	//    params. Shown here as a placeholder:

	// tutorID := r.URL.Query().Get("tutorUserID")
	// tutorEmail := r.URL.Query().Get("tutorEmail")
	// if tutorID == "" || tutorEmail == "" {
	//    http.Error(w, "Missing tutor identity", http.StatusBadRequest)
	//    return
	// }

	// 3. Fetch every student from the "students" collection
	ctx := context.Background()
	iter := a.FirestoreClient.Collection("students").Documents(ctx)
	defer iter.Stop()

	// We'll store matches here
	var matchedStudents []StudentDetailResponse

	for {
		docSnap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error iterating student docs: %v", err)
			http.Error(w, "Failed to read students", http.StatusInternalServerError)
			return
		}

		data := docSnap.Data()
		personal, ok := data["personal"].(map[string]interface{})
		if !ok {
			// No personal subdocument, skip
			continue
		}
		nameVal, ok := personal["name"].(string)
		if !ok {
			// No 'name' field in personal, skip
			continue
		}
		fullName := strings.TrimSpace(nameVal)

		// Check if this student's full name is in the requested set
		if _, found := requestedNames[fullName]; found {
			// We have a match, build the full response object
			studentDetail := StudentDetailResponse{
				ID:       docSnap.Ref.ID,
				Personal: map[string]interface{}{},
				Business: map[string]interface{}{},
			}

			// 3a. Personal data
			if personalData, ok := data["personal"].(map[string]interface{}); ok {
				// You can selectively pick fields or store all
				studentDetail.Personal = personalData
			}

			// 3b. Business data
			if businessData, ok := data["business"].(map[string]interface{}); ok {
				studentDetail.Business = businessData
			}

			// 3c. Subcollections
			// Homework Completion
			homeworkDocs, err := docSnap.Ref.Collection("Homework Completion").Documents(ctx).GetAll()
			if err != nil {
				log.Printf("Error fetching Homework Completion for %s: %v", docSnap.Ref.ID, err)
				studentDetail.HomeworkCompletion = []map[string]interface{}{}
			} else {
				var hwList []map[string]interface{}
				for _, hw := range homeworkDocs {
					hwData := hw.Data()
					hwData["id"] = hw.Ref.ID
					hwList = append(hwList, hwData)
				}
				studentDetail.HomeworkCompletion = hwList
			}

			// Test Data
			testDataDocs, err := docSnap.Ref.Collection("Test Data").Documents(ctx).GetAll()
			if err != nil {
				log.Printf("Error fetching Test Data for %s: %v", docSnap.Ref.ID, err)
				studentDetail.TestData = []map[string]interface{}{}
			} else {
				var tdList []map[string]interface{}
				for _, td := range testDataDocs {
					tdData := td.Data()
					tdData["id"] = td.Ref.ID
					// Optionally parse ACT_Scores or SAT_Scores into arrays as needed:
					if actScores, ok := tdData["ACT_Scores"].(map[string]interface{}); ok {
						tdData["ACT_Scores"] = []interface{}{
							actScores["English"],
							actScores["Math"],
							actScores["Reading"],
							actScores["Science"],
							actScores["ACT_Total"],
						}
					}
					if satScores, ok := tdData["SAT_Scores"].(map[string]interface{}); ok {
						tdData["SAT_Scores"] = []interface{}{
							satScores["EBRW"],
							satScores["Math"],
							satScores["Reading"],
							satScores["Writing"],
							satScores["SAT_Total"],
						}
					}
					tdList = append(tdList, tdData)
				}
				studentDetail.TestData = tdList
			}

			// Test Dates
			testDatesDocs, err := docSnap.Ref.Collection("Test Dates").Documents(ctx).GetAll()
			if err != nil {
				log.Printf("Error fetching Test Dates for %s: %v", docSnap.Ref.ID, err)
				studentDetail.TestDates = []map[string]interface{}{}
			} else {
				var testDates []map[string]interface{}
				for _, tdd := range testDatesDocs {
					tdData := tdd.Data()
					tdData["id"] = tdd.Ref.ID
					testDates = append(testDates, tdData)
				}
				studentDetail.TestDates = testDates
			}

			// Goals
			goalsDocs, err := docSnap.Ref.Collection("Goals").Documents(ctx).GetAll()
			if err != nil {
				log.Printf("Error fetching Goals for %s: %v", docSnap.Ref.ID, err)
				studentDetail.Goals = []map[string]interface{}{}
			} else {
				var goals []map[string]interface{}
				for _, gd := range goalsDocs {
					gData := gd.Data()
					gData["id"] = gd.Ref.ID
					goals = append(goals, gData)
				}
				studentDetail.Goals = goals
			}

			// 4. Add to results
			matchedStudents = append(matchedStudents, studentDetail)
		}
	}

	// 5. Return the matched students as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matchedStudents); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
