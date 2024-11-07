// backend/internal/auth/app.go

package auth

import (
	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"github.com/gorilla/sessions"
)

// App holds the dependencies for the auth package
type App struct {
	Config          *config.Config
	Store           *sessions.CookieStore
	FirestoreClient *firestore.Client
}
