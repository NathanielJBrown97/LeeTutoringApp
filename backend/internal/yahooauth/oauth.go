// backend/internal/yahooauth/oauth.go

package yahooauth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(1 * time.Hour)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Error generating random bytes for state: %v", err)
		return "state" // Fallback to a fixed state in case of error
	}
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	return state
}

func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Log that the OAuthHandler was triggered
	log.Println("Yahoo OAuthHandler triggered")

	// Generate a secure state and store it in a cookie
	state := generateStateOauthCookie(w)

	// Generate the OAuth URL for Yahoo
	url := a.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("nonce", state))

	// Log the generated URL
	log.Println("Redirecting to:", url)

	// Redirect the user to the Yahoo OAuth login page
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
