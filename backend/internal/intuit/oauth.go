package intuit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"

	"golang.org/x/oauth2"
)

// OAuth endpoints for Intuit
var quickBooksEndpoint = oauth2.Endpoint{
	AuthURL:  "https://appcenter.intuit.com/connect/oauth2",
	TokenURL: "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer",
}

// TokenInfo holds the tokens we'll store as a sub-document "intuitoauth".
type TokenInfo struct {
	AccessToken  string    `firestore:"accessToken,omitempty"`
	RefreshToken string    `firestore:"refreshToken,omitempty"`
	TokenType    string    `firestore:"tokenType,omitempty"`
	Expiry       time.Time `firestore:"expiry,omitempty"`
	RealmID      string    `firestore:"realmID,omitempty"`
}

// OAuthService holds the oauth2.Config and the Firestore client.
type OAuthService struct {
	config    *oauth2.Config
	firestore *firestore.Client
}

// NewOAuthService sets up the OAuth config from env vars and holds Firestore ref.
func NewOAuthService(ctx context.Context, fsClient *firestore.Client) (*OAuthService, error) {
	clientID := os.Getenv("INTUIT_CLIENT_ID")
	clientSecret := os.Getenv("INTUIT_CLIENT_SECRET")
	redirectURL := os.Getenv("INTUIT_REDIRECT_URL")
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf("missing INTUIT_CLIENT_ID, INTUIT_CLIENT_SECRET, or INTUIT_REDIRECT_URL")
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     quickBooksEndpoint,
		Scopes: []string{
			"com.intuit.quickbooks.accounting",
		},
	}

	return &OAuthService{
		config:    conf,
		firestore: fsClient,
	}, nil
}

// HandleAuthRedirect initiates the OAuth flow by redirecting the user to Intuit.
func (s *OAuthService) HandleAuthRedirect(w http.ResponseWriter, r *http.Request) {
	state := "someRandomState"
	url := s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback is where Intuit redirects after user authorizes.
// e.g. GET /intuit/callback?code=xxx&realmId=xxx
func (s *OAuthService) HandleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	realmID := query.Get("realmId")
	state := query.Get("state")
	if state != "someRandomState" {
		http.Error(w, "State parameter does not match, or is missing", http.StatusBadRequest)
		return
	}
	if code == "" || realmID == "" {
		http.Error(w, "Missing code or realmId", http.StatusBadRequest)
		return
	}

	// Exchange the code for access/refresh tokens
	token, err := s.config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Error exchanging code: %v\n", err)
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	tokenInfo := TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
		RealmID:      realmID,
	}

	// Store in Firestore -> "intuit/globalTokens" doc, sub-document "intuitoauth"
	if err := s.storeTokens(context.Background(), tokenInfo); err != nil {
		log.Printf("Failed to store tokens: %v\n", err)
		http.Error(w, "Failed to store tokens in Firestore", http.StatusInternalServerError)
		return
	}

	// Redirect user to your chosen success page.
	http.Redirect(w, r, "https://lee-tutoring-webapp.web.app/booking", http.StatusSeeOther)
}

// storeTokens: we store the data under a field "intuitoauth" in the
// doc "intuit/globalTokens". That is effectively a "sub-document" approach.
func (s *OAuthService) storeTokens(ctx context.Context, ti TokenInfo) error {
	// We'll do a merge so we don't overwrite other fields if they exist.
	_, err := s.firestore.Collection("intuit").Doc("globalTokens").
		Set(ctx, map[string]interface{}{
			"intuitoauth": ti,
		}, firestore.MergeAll)
	return err
}

// retrieveTokens loads from sub-document "intuitoauth"
func (s *OAuthService) retrieveTokens(ctx context.Context) (*TokenInfo, error) {
	docSnap, err := s.firestore.Collection("intuit").Doc("globalTokens").Get(ctx)
	if err != nil {
		return nil, err
	}
	var temp struct {
		Intuitoauth TokenInfo `firestore:"intuitoauth"`
	}
	if err := docSnap.DataTo(&temp); err != nil {
		return nil, err
	}
	return &temp.Intuitoauth, nil
}

// Token returns a fresh or valid token, refreshing if expired
func (s *OAuthService) Token(ctx context.Context) (*oauth2.Token, error) {
	ti, err := s.retrieveTokens(ctx)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{
		AccessToken:  ti.AccessToken,
		RefreshToken: ti.RefreshToken,
		TokenType:    ti.TokenType,
		Expiry:       ti.Expiry,
	}
	ts := s.config.TokenSource(ctx, tok)
	newTok, err := ts.Token()
	if err != nil {
		return nil, err
	}

	// If it changed, store the updated version
	if newTok.AccessToken != ti.AccessToken {
		updated := TokenInfo{
			AccessToken:  newTok.AccessToken,
			RefreshToken: newTok.RefreshToken,
			TokenType:    newTok.TokenType,
			Expiry:       newTok.Expiry,
			RealmID:      ti.RealmID,
		}
		if err := s.storeTokens(ctx, updated); err != nil {
			return nil, fmt.Errorf("failed to store updated token: %w", err)
		}
	}
	return newTok, nil
}
