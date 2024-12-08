// backend/internal/appleauth/oauth.go

package appleauth

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Log that the OAuthHandler was triggered
	log.Println("Apple OAuthHandler triggered")

	// Generate the OAuth URL for Apple
	url := a.OAuthConfig.AuthCodeURL("state",
		oauth2.SetAuthURLParam("response_type", "code id_token"),
		oauth2.SetAuthURLParam("response_mode", "form_post"),
		oauth2.SetAuthURLParam("scope", "name email"),
		oauth2.SetAuthURLParam("nonce", "random_nonce"), // Replace with actual nonce generation
	)

	// Log the generated URL
	log.Println("Redirecting to:", url)

	// Redirect the user to the Apple OAuth login page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
