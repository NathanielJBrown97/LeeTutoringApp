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
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/yahooauth"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"

	"cloud.google.com/go/firestore"
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
		Scopes:       []string{"email", "profile"},
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

	// Set up the HTTP server and routes using gorilla/mux
	r := mux.NewRouter()

	// Use the AuthMiddleware for protected routes
	authMiddleware := middleware.AuthMiddleware(secretKey)

	// API routes with CORS and OPTIONS handling

	// Dashboard route
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
