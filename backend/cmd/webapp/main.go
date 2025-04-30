// backend/main.go

package main

import (
	"context"
	"log"
	"net/http"
	"os"

	firestoreupdater "github.com/NathanielJBrown97/LeeTutoringApp/cmd/firestoreupdater"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/appleauth"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/auth"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/dashboard"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/facebookauth"
	googleauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/googleauth"
	microsoftauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/microsoftauth"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/middleware"
	parentpkg "github.com/NathanielJBrown97/LeeTutoringApp/internal/parent"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/tutordashboard"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/yahooauth"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"

	"cloud.google.com/go/firestore"
	intuitoauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/intuit"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(context.Background(), cfg.GOOGLE_CLOUD_PROJECT)
	if err != nil {
		log.Fatalf("Error initializing Firestore: %v", err)
	}
	defer firestoreClient.Close()

	// Secret key for JWT
	secretKey := cfg.SESSION_SECRET

	// Google OAuth2 configuration
	googleConf := &oauth2.Config{
		ClientID:     cfg.GOOGLE_CLIENT_ID,
		ClientSecret: cfg.GOOGLE_CLIENT_SECRET,
		RedirectURL:  cfg.GOOGLE_REDIRECT_URL,
		Scopes:       []string{"email", "profile", "https://www.googleapis.com/auth/calendar.readonly"},
		Endpoint:     google.Endpoint,
	}

	googleApp := googleauth.App{
		Config:          cfg,
		OAuthConfig:     googleConf,
		FirestoreClient: firestoreClient,
		SecretKey:       secretKey,
	}

	// Microsoft OAuth2 configuration
	microsoftOauthConfig := &oauth2.Config{
		ClientID:     cfg.MICROSOFT_CLIENT_ID,
		ClientSecret: cfg.MICROSOFT_CLIENT_SECRET,
		RedirectURL:  cfg.MICROSOFT_REDIRECT_URL,
		Scopes:       []string{"openid", "email", "profile", "User.Read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}

	microsoftApp := &microsoftauth.App{
		Config:          cfg,
		OAuthConfig:     microsoftOauthConfig,
		FirestoreClient: firestoreClient,
		SecretKey:       secretKey,
	}

	// Yahoo OAuth2 configuration
	yahooConf := &oauth2.Config{
		ClientID:     cfg.YAHOO_CLIENT_ID,
		ClientSecret: cfg.YAHOO_CLIENT_SECRET,
		RedirectURL:  cfg.YAHOO_REDIRECT_URL,
		Scopes:       []string{"openid"},
		Endpoint:     yahooauth.YahooEndpoint,
	}

	yahooApp := yahooauth.App{
		Config:          cfg,
		OAuthConfig:     yahooConf,
		FirestoreClient: firestoreClient,
		SecretKey:       secretKey,
	}

	// Facebook OAuth2 configuration
	facebookConf := &oauth2.Config{
		ClientID:     cfg.FACEBOOK_CLIENT_ID,
		ClientSecret: cfg.FACEBOOK_CLIENT_SECRET,
		RedirectURL:  cfg.FACEBOOK_REDIRECT_URL,
		Scopes:       []string{"public_profile", "email"},
		Endpoint:     facebook.Endpoint,
	}

	facebookApp := facebookauth.App{
		Config:          cfg,
		OAuthConfig:     facebookConf,
		FirestoreClient: firestoreClient,
		SecretKey:       secretKey,
	}

	// Apple OAuth2 configuration
	appleConf := &oauth2.Config{
		ClientID:    cfg.APPLE_CLIENT_ID,
		RedirectURL: cfg.APPLE_REDIRECT_URL,
		Scopes:      []string{"name", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
		},
	}

	appleApp := appleauth.App{
		Config:          cfg,
		OAuthConfig:     appleConf,
		FirestoreClient: firestoreClient,
		SecretKey:       secretKey,
	}

	// Initialize parent App
	parentApp := parentpkg.App{
		Config:          cfg,
		FirestoreClient: firestoreClient,
	}

	// Initialize dashboard App
	dashboardApp := dashboard.App{
		Config:          cfg,
		FirestoreClient: firestoreClient,
	}

	// Initialize auth App
	authApp := auth.App{
		Config:          cfg,
		FirestoreClient: firestoreClient,
	}

	// Initialize FirestoreUpdater App
	firestoreUpdaterApp := firestoreupdater.App{
		Config:          cfg,
		FirestoreClient: firestoreClient,
	}

	// Create an instance of the tutor dashboard app.
	tutorDashboardApp := tutordashboard.App{
		FirestoreClient: firestoreClient,
		// ... initialize other fields if necessary
	}
	// Initialize Intuit OAuth Services
	intuitOAuthSvc, err := intuitoauth.NewOAuthService(context.Background(), firestoreClient)
	if err != nil {
		log.Fatalf("Failed to init Intuit OAuth Service: %v", err)
	}

	// Set up the HTTP server and routes using gorilla/mux
	r := mux.NewRouter()

	// Use the AuthMiddleware for protected routes
	authMiddleware := middleware.AuthMiddleware(secretKey)

	// TUTOR DASHBOARD HANDLERS
	// TUTOR TOOLS - Assign Homework route
	r.HandleFunc("/api/tutor/assign-homework", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Wrap with your auth middleware if needed.
		authMiddleware(http.HandlerFunc(tutordashboard.AssignHomeworkHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// Tutor Calendar Events route
	r.HandleFunc("/api/tutor/calendar-events", func(w http.ResponseWriter, r *http.Request) {
		authMiddleware(http.HandlerFunc(tutorDashboardApp.CalendarEventsHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// associate students route
	r.HandleFunc("/api/tutor/associate-students", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Wrap with the auth middleware to ensure only authenticated tutors can trigger it.
		authMiddleware(http.HandlerFunc(tutordashboard.AssociateStudentsHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Tutor Profile Route
	r.HandleFunc("/api/tutor/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.FetchTutorProfileHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Tutor Student Detail route - returns detailed student info only if the student is associated with the tutor.
	r.HandleFunc("/api/tutor/students/{student_id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutorDashboardApp.TutorStudentDetailHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Tutor get students by name
	r.HandleFunc("/api/tutor/fetch-students-by-names", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Wrap with authMiddleware so only authenticated tutors can access
		authMiddleware(http.HandlerFunc(tutorDashboardApp.FetchStudentsByNamesHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// This endpoint returns the list of associated students (IDs and optionally names) for the tutor.
	r.HandleFunc("/api/tutor/fetch-associated-students", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.FetchAssociatedStudentsHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// TUTOR TOOLS
	// edit personal details
	r.HandleFunc("/api/tutor/edit-personal-details", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.EditPersonalDetailsHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// get personal detalis:
	r.HandleFunc("/api/tutor/get-personal-details", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.GetPersonalDetailsHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Get Business Details
	r.HandleFunc("/api/tutor/get-business-details", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.GetBusinessDetailsHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// edit business data
	r.HandleFunc("/api/tutor/edit-business-details", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.EditBusinessDetailsHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// delete test data
	r.HandleFunc("/api/tutor/delete-test-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.DeleteTestDataHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// Create Test Data
	r.HandleFunc("/api/tutor/create-test-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.CreateTestDataHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// Create Homework Completion
	r.HandleFunc("/api/tutor/create-homework-completion", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.CreateHomeworkCompletionHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// edit Homework Completion
	r.HandleFunc("/api/tutor/edit-homework-completion", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.EditHomeworkCompletionHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")
	// delete Homework Completion
	r.HandleFunc("/api/tutor/delete-homework-completion", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.DeleteHomeworkCompletionHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// Create Test Data
	r.HandleFunc("/api/tutor/edit-test-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.EditTestDataHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// create goals
	r.HandleFunc("/api/tutor/create-goal", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.CreateGoalHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// delete goals
	r.HandleFunc("/api/tutor/delete-goal", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.DeleteGoalHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// Edit Test Data Notes
	r.HandleFunc("/api/tutor/edit-test-dates-notes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(tutordashboard.EditTestDatesNotesHandler(firestoreClient))).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// PARENT Dashboard route
	r.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.Handler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Associated students route
	r.HandleFunc("/api/associated-students", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.AssociatedStudentsHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Student detail route
	r.HandleFunc("/api/students/{student_id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.StudentDetailHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Parent routes
	r.HandleFunc("/api/submitStudentIDs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(parentApp.StudentIntakeHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/confirmLinkStudents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(parentApp.ConfirmLinkStudentsHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/attemptAutomaticAssociation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(parentApp.AttemptAutomaticAssociation)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// api request for hours/and balance
	r.HandleFunc("/api/dashboard/total-hours-and-balance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.TotalHoursAndBalanceHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// api to update STUDENTS lifetime hours
	r.HandleFunc("/api/students/{student_id}/update_student_lifetime_hours", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.UpdateStudentLifetimeHoursHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// api endpoint, updates parents used hours - iteratively updates each associated students lifetiem hours
	r.HandleFunc("/api/parent/update_used_hours", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.UpdateParentUsedHoursHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// gets all parent data related to invoices, payments, voids, ect.
	r.HandleFunc("/api/parent/invoices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.GetParentInvoicesHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Auth status route
	r.Handle("/api/auth/status", authMiddleware(http.HandlerFunc(authApp.StatusHandler))).Methods("GET", "OPTIONS")

	// ParentHandler route
	r.HandleFunc("/api/parent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.ParentHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/updateInvoiceEmail", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Wrap with your auth middleware if required
		authMiddleware(http.HandlerFunc(dashboardApp.UpdateInvoiceEmailHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/getInvoiceEmail", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Wrap with your auth middleware if required
		authMiddleware(http.HandlerFunc(dashboardApp.GetInvoiceEmailHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// intuit related handlers including oauth
	r.HandleFunc("/internal/intuit/oauth", func(w http.ResponseWriter, r *http.Request) {
		intuitOAuthSvc.HandleAuthRedirect(w, r)
	}).Methods("GET")

	r.HandleFunc("/internal/intuit/callback", func(w http.ResponseWriter, r *http.Request) {
		intuitOAuthSvc.HandleCallback(w, r)
	}).Methods("GET")

	// intuit webhook for keeping invoices up to date
	r.HandleFunc("/internal/intuit/webhook", intuitOAuthSvc.HandleWebhook).Methods("POST")
	// intuit daily poll endpoint (GCP Cloud Scheduler)
	r.HandleFunc("/internal/intuit/daily-poll", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Just call the new poll handler method you added in webhook.go
		intuitOAuthSvc.HandleDailyPoll(w, r)
	}).Methods("GET", "OPTIONS")
	// OAUTH HANDLERS

	// Google OAuth handlers
	r.HandleFunc("/internal/googleauth/oauth", googleApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/googleauth/callback", googleApp.OAuthCallbackHandler).Methods("GET")

	// Microsoft OAuth handlers
	r.HandleFunc("/internal/microsoftauth/oauth", microsoftApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/microsoftauth/callback", microsoftApp.OAuthCallbackHandler).Methods("GET")
	// **Microsoft's ProfilePictureHandler**
	r.HandleFunc("/internal/microsoftauth/profile-picture", microsoftApp.ProfilePictureHandler).Methods("GET")

	// Yahoo OAuth handlers
	r.HandleFunc("/internal/yahooauth/oauth", yahooApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/yahooauth/callback", yahooApp.OAuthCallbackHandler).Methods("GET")

	// Facebook OAuth handlers
	r.HandleFunc("/internal/facebookauth/oauth", facebookApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/facebookauth/callback", facebookApp.OAuthCallbackHandler).Methods("GET")

	// Apple OAuth handlers
	r.HandleFunc("/internal/appleauth/oauth", appleApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/appleauth/callback", appleApp.OAuthCallbackHandler).Methods("POST")

	// Firestore Updater Routes with CORS and OPTIONS handling

	// Initialize New Student
	r.HandleFunc("/cmd/firestoreupdater/initializeNewStudent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		firestoreupdater.InitializeNewStudent(w, r)
	}).Methods("POST", "OPTIONS")

	// Homework Completion
	r.HandleFunc("/cmd/firestoreupdater/homeworkCompletion", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		firestoreUpdaterApp.UpdateHomeworkCompletionHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Test Data
	r.HandleFunc("/cmd/firestoreupdater/testData", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		firestoreUpdaterApp.UpdateTestDataHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Test Dates Trigger
	r.HandleFunc("/cmd/firestoreupdater/testDatesTrigger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		firestoreUpdaterApp.UpdateTestDatesHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Update Goals
	r.HandleFunc("/cmd/firestoreupdater/updateGoals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		firestoreUpdaterApp.UpdateGoalsHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Profile Data
	r.HandleFunc("/cmd/firestoreupdater/profileData", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Handle preflight request
			w.WriteHeader(http.StatusNoContent)
			return
		}
		firestoreUpdaterApp.UpdateProfileHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Use CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://lee-tutoring-webapp.web.app", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            false, // Disable debug mode in production
	})
	handler := c.Handler(r)

	// Get the port from the environment variable PORT (GCP uses this)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the HTTP server
	log.Printf("Server started on port %s", port)
	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
