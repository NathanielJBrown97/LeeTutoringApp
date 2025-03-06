// File: calendar.go
package tutordashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Tutor represents the structure of a tutor document in Firestore.
type Tutor struct {
	UserID       string    `firestore:"user_id"`
	Email        string    `firestore:"email"`
	AccessToken  string    `firestore:"access_token"`
	RefreshToken string    `firestore:"refresh_token"`
	Expiry       time.Time `firestore:"expiry"`
	Name         string    `firestore:"name"`
	Picture      string    `firestore:"picture"`
	CalendarID   string    `firestore:"calendar_id,omitempty"`
}

// getTutor retrieves the tutor document from the "tutors" collection using the userID.
func getTutor(ctx context.Context, client *firestore.Client, userID string) (*Tutor, error) {
	doc, err := client.Collection("tutors").Doc(userID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var tutor Tutor
	if err := doc.DataTo(&tutor); err != nil {
		return nil, err
	}
	return &tutor, nil
}

// getTutorCalendarID returns the specific calendar ID for a tutor.
// If not set, it defaults to "primary".
func getTutorCalendarID(tutor *Tutor) string {
	if tutor.CalendarID != "" {
		return tutor.CalendarID
	}
	return "primary"
}

// CalendarEventsHandler handles HTTP requests to fetch today's calendar events.
// It is now a method on the App struct.
func (app *App) CalendarEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Assume the tutor's userID is passed as a query parameter.
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Fetch tutor credentials from Firestore using app.FirestoreClient.
	tutor, err := getTutor(ctx, app.FirestoreClient, userID)
	if err != nil {
		log.Printf("Error fetching tutor data: %v", err)
		http.Error(w, "Tutor not found", http.StatusNotFound)
		return
	}

	// Construct an OAuth2 token from the Firestore credentials.
	token := &oauth2.Token{
		AccessToken:  tutor.AccessToken,
		RefreshToken: tutor.RefreshToken,
		Expiry:       tutor.Expiry,
	}

	// Check token validity.
	if !token.Valid() {
		log.Println("Access token expired. Refresh logic needs to be implemented.")
		// In production, you should refresh the token or return an error.
		// For now, we'll continue using the token.
	}

	// Create a token source using the tutor's token.
	tokenSource := oauth2.StaticTokenSource(token)

	// Create a Google Calendar service using the token.
	calendarService, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Printf("Error creating calendar service: %v", err)
		http.Error(w, "Failed to create calendar service", http.StatusInternalServerError)
		return
	}

	// Determine the calendar ID to use.
	calendarID := getTutorCalendarID(tutor)

	// Define the time range for "today".
	nowTime := time.Now()
	year, month, day := nowTime.Date()
	location := nowTime.Location()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Query events for today from the selected calendar.
	events, err := calendarService.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(startOfDay.Format(time.RFC3339)).
		TimeMax(endOfDay.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Printf("Error retrieving events: %v", err)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Return the events as JSON.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
