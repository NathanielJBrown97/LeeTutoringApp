// backend/main.go

package main

import (
	"log"
	"net/http"
	"os"

	firestoreupdater "github.com/NathanielJBrown97/LeeTutoringApp/cmd/firestoreupdater"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/auth"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/dashboard"
	googleauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/googleauth"
	microsoftauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/microsoftauth"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/middleware"
	parentpkg "github.com/NathanielJBrown97/LeeTutoringApp/internal/parent"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize Firestore client
	firestoreClient, err := config.InitializeFirestore(cfg)
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
	microsoftConf := &oauth2.Config{
		ClientID:     cfg.MICROSOFT_CLIENT_ID,
		ClientSecret: cfg.MICROSOFT_CLIENT_SECRET,
		RedirectURL:  cfg.MICROSOFT_REDIRECT_URL,
		Scopes:       []string{"openid", "profile", "email", "offline_access"},
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}

	microsoftApp := microsoftauth.App{
		Config:          cfg,
		OAuthConfig:     microsoftConf,
		FirestoreClient: firestoreClient,
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

	// API routes

	// Dashboard route with OPTIONS handling
	r.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.Handler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Associated students route with OPTIONS handling
	r.HandleFunc("/api/associated-students", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.AssociatedStudentsHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Student detail route with OPTIONS handling
	r.HandleFunc("/api/students/{student_id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(dashboardApp.StudentDetailHandler)).ServeHTTP(w, r)
	}).Methods("GET", "OPTIONS")

	// Parent routes
	r.HandleFunc("/api/submitStudentIDs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(parentApp.StudentIntakeHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/confirmLinkStudents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(parentApp.ConfirmLinkStudentsHandler)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/attemptAutomaticAssociation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authMiddleware(http.HandlerFunc(parentApp.AttemptAutomaticAssociation)).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	r.Handle("/api/auth/status", authMiddleware(http.HandlerFunc(authApp.StatusHandler))).Methods("GET", "OPTIONS")

	// OAuth handlers
	r.HandleFunc("/internal/googleauth/oauth", googleApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/googleauth/callback", googleApp.OAuthCallbackHandler).Methods("GET")
	r.HandleFunc("/internal/microsoftauth/oauth", microsoftApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/microsoftauth/callback", microsoftApp.OAuthCallbackHandler).Methods("GET")

	// Firestore Updater Routes

	//Firestore updater -- HOMEWORK COMPLETION
	r.HandleFunc("/cmd/firestoreupdater/homeworkCompletion", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Call the handler
		firestoreUpdaterApp.UpdateHomeworkCompletionHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Firestore updater -- TEST DATA
	r.HandleFunc("/cmd/firestoreupdater/testData", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Call the handler
		firestoreUpdaterApp.UpdateTestDataHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Firestore updater -- TEST DATES TRIGGER
	r.HandleFunc("/cmd/firestoreupdater/testDatesTrigger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Call the handler
		firestoreUpdaterApp.UpdateTestDatesHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Firestore updater -- GOALS TRIGGER
	r.HandleFunc("/cmd/firestoreupdater/updateGoals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Call the handler
		firestoreUpdaterApp.UpdateGoalsHandler(w, r)
	}).Methods("POST", "OPTIONS")

	// Use CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://lee-tutoring-webapp.web.app", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true, // Enable debug mode to log CORS issues
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
