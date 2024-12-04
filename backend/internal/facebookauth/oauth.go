package facebookauth

import (
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Log that the OAuthHandler was triggered
	log.Println("Facebook OAuthHandler triggered")

	// Generate the OAuth URL for Facebook
	url := a.OAuthConfig.AuthCodeURL("state", oauth2.SetAuthURLParam("scope", "public_profile,email"))

	// Log the generated URL
	log.Println("Redirecting to:", url)

	// Redirect the user to the Facebook OAuth login page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
