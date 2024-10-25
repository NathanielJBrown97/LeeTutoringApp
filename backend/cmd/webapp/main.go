// backend/main.go

package main

import (
	"log"
	"net/http"

	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/dashboard"
	googleauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/googleauth"
	microsoftauth "github.com/NathanielJBrown97/LeeTutoringApp/internal/microsoftauth"
	parentpkg "github.com/NathanielJBrown97/LeeTutoringApp/internal/parent"
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
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // Adjust as needed
		// Secure:   true, // Uncomment when using HTTPS
	}

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
		Store:           store,
	}

	// Microsoft OAuth2 configuration
	microsoftConf := &oauth2.Config{
		ClientID:     cfg.MICROSOFT_CLIENT_ID,
		ClientSecret: cfg.MICROSOFT_CLIENT_SECRET,
		RedirectURL:  cfg.MICROSOFT_REDIRECT_URL,
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

	// Set up the HTTP server and routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/dashboard", dashboardApp.Handler)
	mux.HandleFunc("/api/submitStudentIDs", parentApp.StudentIntakeHandler)
	mux.HandleFunc("/api/confirmLinkStudents", parentApp.ConfirmLinkStudentsHandler)

	// OAuth handlers
	mux.HandleFunc("/internal/googleauth/oauth", googleApp.OAuthHandler)
	mux.HandleFunc("/internal/googleauth/callback", googleApp.OAuthCallbackHandler)
	mux.HandleFunc("/internal/microsoftauth/oauth", microsoftApp.OAuthHandler)
	mux.HandleFunc("/internal/microsoftauth/callback", microsoftApp.OAuthCallbackHandler)

	// Optionally serve static files (in production)
	// Commented out during development
	// fs := http.FileServer(http.Dir("./frontend/build"))
	// mux.Handle("/", fs)

	// Use CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)

	// Start the HTTP server
	log.Printf("Server started on http://localhost:%s", cfg.PORT)
	err = http.ListenAndServe(":"+cfg.PORT, handler)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
