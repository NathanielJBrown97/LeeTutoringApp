// backend/internal/dashboard/app.go

package dashboard

import (
	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/backend/internal/config"
	"github.com/gorilla/sessions"
)

// App holds the dependencies for the dashboard package
type App struct {
	Config          *config.Config
	FirestoreClient *firestore.Client
	Store           *sessions.CookieStore
}
