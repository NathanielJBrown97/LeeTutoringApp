// backend/internal/microsoftauth/app.go

package microsoftauth

import (
	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"golang.org/x/oauth2"
)

// App holds the dependencies for the microsoftauth package
type App struct {
	Config          *config.Config
	OAuthConfig     *oauth2.Config
	FirestoreClient *firestore.Client
	SecretKey       string
}
