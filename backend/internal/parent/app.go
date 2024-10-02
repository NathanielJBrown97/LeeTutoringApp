// backend/internal/parent/app.go

package parent

import (
	"cloud.google.com/go/firestore"

	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"github.com/gorilla/sessions"
)

// App holds the dependencies for the parent package
type App struct {
	Config          *config.Config
	FirestoreClient *firestore.Client
	Store           *sessions.CookieStore
}
