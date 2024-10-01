// backend/internal/microsoftauth/oauth.go

package microsoftauth

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Log that the OAuthHandler was triggered
	log.Println("OAuthHandler triggered")

	// Generate the OAuth URL for Microsoft
	url := a.OAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)

	// Log the generated URL
	log.Println("Redirecting to:", url)

	// Redirect the user to the Microsoft OAuth login page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
