// backend/main.go

package main

import (
	"log"
	"net/http"

	"github.com/NathanielJBrown97/LeeTutoringApp/backend/internal/config"
	"github.com/NathanielJBrown97/LeeTutoringApp/backend/internal/dashboard"
	googleauth "github.com/NathanielJBrown97/LeeTutoringApp/backend/internal/googleauth"
	microsoftauth "github.com/NathanielJBrown97/LeeTutoringApp/backend/internal/microsoftauth"
	parentpkg "github.com/NathanielJBrown97/LeeTutoringApp/backend/internal/parent"
	"github.com/gorilla/sessions"
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
	// Optionally, set session options
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
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
		Scopes:       []string{"openid", "profile", "email", "offline_access", "https://graph.microsoft.com/User.Read"},
		Endpoint:     microsoft.AzureADEndpoint("common"), // Replace "common" with your tenant ID if needed
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

	// Google OAuth handlers
	mux.HandleFunc("/", googleApp.LoginHandler)
	mux.HandleFunc("/internal/googleauth/oauth", googleApp.OAuthHandler)
	mux.HandleFunc("/internal/googleauth/callback", googleApp.OAuthCallbackHandler)

	// Microsoft OAuth handlers
	mux.HandleFunc("/internal/microsoftauth/oauth", microsoftApp.OAuthHandler)
	mux.HandleFunc("/internal/microsoftauth/callback", microsoftApp.OAuthCallbackHandler)

	// Parent dashboard handler
	mux.HandleFunc("/parentdashboard", dashboardApp.Handler) // Updated line

	// Parent intake page
	mux.HandleFunc("/parentintake", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("ParentIntake received a %s request\n", r.Method)
		log.Printf("Headers: %v\n", r.Header)
		http.ServeFile(w, r, "parentintake.html")
	})

	// Parental intake handling routes
	mux.HandleFunc("/submitStudentIDs", parentApp.StudentIntakeHandler)
	mux.HandleFunc("/confirmLinkStudents", parentApp.ConfirmLinkStudentsHandler) // Ensure this handler is defined similarly

	// Start the HTTP server
	log.Printf("Server started on http://localhost:%s", cfg.PORT)
	err = http.ListenAndServe(":"+cfg.PORT, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
