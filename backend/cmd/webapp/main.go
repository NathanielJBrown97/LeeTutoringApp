// backend/main.go

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/NathanielJBrown97/LeeTutoringApp/internal/auth"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/dashboard"
	googleauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/googleauth"
	microsoftauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/microsoftauth"
	parentpkg "github.com/NathanielJBrown97/LeeTutoringApp/internal/parent"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
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

	// Initialize session store with secret from config
	store := sessions.NewCookieStore([]byte(cfg.SESSION_SECRET))
	// Set session options
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,             // 7 days
		HttpOnly: true,                  // Prevents JavaScript access to cookies
		SameSite: http.SameSiteNoneMode, // Allows cross-site cookies
		Secure:   true,                  // Ensures cookies are only sent over HTTPS
	}

	// Google OAuth2 configuration with AccessTypeOffline and PromptConsent
	googleConf := &oauth2.Config{
		ClientID:     cfg.GOOGLE_CLIENT_ID,
		ClientSecret: cfg.GOOGLE_CLIENT_SECRET,
		RedirectURL:  cfg.GOOGLE_REDIRECT_URL, // Should be set to your backend's callback URL
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	googleApp := googleauth.App{
		Config:          cfg,
		OAuthConfig:     googleConf,
		FirestoreClient: firestoreClient,
		Store:           store,
	}

	// Microsoft OAuth2 configuration
	microsoftConf := &oauth2.Config{
		ClientID:     cfg.MICROSOFT_CLIENT_ID,
		ClientSecret: cfg.MICROSOFT_CLIENT_SECRET,
		RedirectURL:  cfg.MICROSOFT_REDIRECT_URL, // Should be set to your backend's callback URL
		Scopes:       []string{"openid", "profile", "email", "offline_access"},
		Endpoint:     microsoft.AzureADEndpoint("common"), // Use "common" to support all account types
	}

	microsoftApp := microsoftauth.App{
		Config:          cfg,
		OAuthConfig:     microsoftConf,
		FirestoreClient: firestoreClient,
		Store:           store,
	}

	// Initialize parent App
	parentApp := parentpkg.App{
		Config:          cfg,
		FirestoreClient: firestoreClient,
		Store:           store,
	}

	// Initialize dashboard App
	dashboardApp := dashboard.App{
		Config:          cfg,
		FirestoreClient: firestoreClient,
		Store:           store,
	}

	// Initialize auth App
	authApp := auth.App{
		Config:          cfg,
		Store:           store,
		FirestoreClient: firestoreClient,
	}

	// Set up the HTTP server and routes using gorilla/mux
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/dashboard", dashboardApp.Handler).Methods("GET")
	r.HandleFunc("/api/submitStudentIDs", parentApp.StudentIntakeHandler).Methods("POST")
	r.HandleFunc("/api/confirmLinkStudents", parentApp.ConfirmLinkStudentsHandler).Methods("POST")
	r.HandleFunc("/api/auth/status", authApp.StatusHandler).Methods("GET") // Auth Status Endpoint

	// New endpoints for accessing student data
	r.HandleFunc("/api/students", dashboardApp.StudentsHandler).Methods("GET")
	r.HandleFunc("/api/students/{student_id}", dashboardApp.StudentDetailHandler).Methods("GET")

	// OAuth handlers
	r.HandleFunc("/internal/googleauth/oauth", googleApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/googleauth/callback", googleApp.OAuthCallbackHandler).Methods("GET")
	r.HandleFunc("/internal/microsoftauth/oauth", microsoftApp.OAuthHandler).Methods("GET")
	r.HandleFunc("/internal/microsoftauth/callback", microsoftApp.OAuthCallbackHandler).Methods("GET")

	// Use CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://lee-tutoring-webapp.web.app"}, // Your frontend domain
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
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
