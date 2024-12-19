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
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, &cookie)

	return state
}

func (a *App) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Yahoo OAuthHandler triggered")

	state := generateStateOauthCookie(w)
	url := a.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("nonce", state))

	log.Println("Redirecting to Yahoo OAuth URL:", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
