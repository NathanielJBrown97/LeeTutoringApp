// backend/internal/googleauth/oauth.go

package googleauth

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Log that the OAuthHandler was triggered
	log.Println("OAuthHandler triggered")

	// Generate the OAuth URL for Google
	url := a.OAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)

	// Log the generated URL
	log.Println("Redirecting to:", url)

	// Redirect the user to the Google OAuth login page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
