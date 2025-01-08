package appleauth

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

// OAuthHandler initiates the Apple Sign In flow with "code id_token" so we can link users in callback.
func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Apple OAuthHandler triggered")

	// Request "code id_token", "form_post", plus "name email" so Apple can pass user data on first login.
	url := a.OAuthConfig.AuthCodeURL("state",
		oauth2.SetAuthURLParam("response_type", "code id_token"),
		oauth2.SetAuthURLParam("response_mode", "form_post"),
		oauth2.SetAuthURLParam("scope", "name email"),
		// Typically you also set a random nonce if you're verifying ID token integrity
		oauth2.SetAuthURLParam("nonce", "random_nonce"),
	)

	log.Println("Redirecting to:", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
