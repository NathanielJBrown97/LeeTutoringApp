// backend/internal/yahooauth/app.go

package yahooauth

import (
	"cloud.google.com/go/firestore"
	"github.com/NathanielJBrown97/LeeTutoringApp/internal/config"
	"golang.org/x/oauth2"
)

// Yahoo OAuth2 Endpoint
var YahooEndpoint = oauth2.Endpoint{
	AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
	TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
}

// App holds the dependencies for the yahooauth package
type App struct {
	Config          *config.Config
	OAuthConfig     *oauth2.Config
	FirestoreClient *firestore.Client
	SecretKey       string
}
