package intuit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/firestore"

	"golang.org/x/oauth2"
)

// QuickBooks sandbox endpoints.
// If you move to production, remove "sandbox-" from the base domain.
var quickBooksEndpoint = oauth2.Endpoint{
	AuthURL:  "https://appcenter.intuit.com/connect/oauth2",
	TokenURL: "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer",
}

// IntuitIntegration represents the sub-document you'll store in Firestore
// inside the parent's document. You can add whatever fields you need.
type IntuitIntegration struct {
	RealmID      string    `firestore:"realmID,omitempty"`
	AccessToken  string    `firestore:"accessToken,omitempty"`
	RefreshToken string    `firestore:"refreshToken,omitempty"`
	TokenType    string    `firestore:"tokenType,omitempty"`
	Expiry       time.Time `firestore:"expiry,omitempty"`

	// Potential fields for hours data (or you may store them separately).
	HoursPurchased float64 `firestore:"hoursPurchased,omitempty"`
	HoursUsed      float64 `firestore:"hoursUsed,omitempty"`
	HoursRemaining float64 `firestore:"hoursRemaining,omitempty"`
}

// OAuthService holds references needed for our Intuit integration.
type OAuthService struct {
	config    *oauth2.Config
	firestore *firestore.Client
}

// NewOAuthService sets up the Intuit OAuth config and stores a Firestore client reference.
// Adjust as needed for your app’s structure.
func NewOAuthService(ctx context.Context, fsClient *firestore.Client) (*OAuthService, error) {
	clientID := os.Getenv("INTUIT_CLIENT_ID")
	clientSecret := os.Getenv("INTUIT_CLIENT_SECRET")
	redirectURL := os.Getenv("INTUIT_REDIRECT_URL")
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf("missing one of INTUIT_CLIENT_ID, INTUIT_CLIENT_SECRET, INTUIT_REDIRECT_URL env vars")
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     quickBooksEndpoint,
		Scopes: []string{
			// Typically for QuickBooks:
			"com.intuit.quickbooks.accounting",
		},
	}

	svc := &OAuthService{
		config:    conf,
		firestore: fsClient,
	}
	return svc, nil
}

// HandleAuthRedirect is hit when you want to start the OAuth flow.
// For example: GET /intuit/auth?parentID=<some-uid>
func (s *OAuthService) HandleAuthRedirect(w http.ResponseWriter, r *http.Request) {
	parentID := r.URL.Query().Get("parentID")
	if parentID == "" {
		http.Error(w, "missing parentID", http.StatusBadRequest)
		return
	}

	// Typically you'd store this state in a session or sign it with a secret to prevent CSRF.
	state := parentID

	authURL := s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleCallback is the endpoint Intuit will redirect to after user authorizes.
// Ex: GET /intuit/callback?state=<parentID>&code=<code>&realmId=<realmId>
func (s *OAuthService) HandleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	realmID := query.Get("realmId")
	state := query.Get("state") // we used parentID as the state
	if code == "" || realmID == "" || state == "" {
		http.Error(w, "invalid callback params", http.StatusBadRequest)
		return
	}

	parentID := state // We assume state is the parent ID you passed earlier.

	token, err := s.config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Token exchange error: %v\n", err)
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Save tokens + realmID into Firestore
	intuitData := IntuitIntegration{
		RealmID:      realmID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}

	err = s.storeIntuitData(context.Background(), parentID, intuitData)
	if err != nil {
		log.Printf("Failed to store Intuit data: %v\n", err)
		http.Error(w, "failed to store tokens in Firestore", http.StatusInternalServerError)
		return
	}

	// For demonstration, you might fetch invoices immediately to store hours purchased
	// or schedule a background job. We'll do it inline just to show how:
	err = s.fetchAndStoreHoursForParent(context.Background(), parentID)
	if err != nil {
		log.Printf("Failed to fetch hours: %v\n", err)
		// Not a fatal error for OAuth—just log it.
		// Could store partial data or handle differently.
	}

	// Redirect the user back to your frontend parent dashboard  I NEED TO CHANGE THIS PARENT ID IS GOING TO GOOF
	http.Redirect(w, r, "https://lee-tutoring-webapp.web.app/parentdashboard?studentID="+parentID, http.StatusSeeOther)
}

// fetchAndStoreHoursForParent uses the parent's tokens in Firestore, calls QuickBooks sandbox to retrieve invoice data,
// calculates the "hours purchased", then updates Firestore with the new hours.
func (s *OAuthService) fetchAndStoreHoursForParent(ctx context.Context, parentID string) error {
	// 1. Get the parent’s IntuitIntegration sub-document
	docSnap, err := s.firestore.Collection("parents").Doc(parentID).Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get parent doc: %w", err)
	}

	var stored IntuitIntegration
	if err := docSnap.DataTo(&stored); err != nil {
		return fmt.Errorf("failed to parse doc data: %w", err)
	}

	// 2. Build a token object
	token := &oauth2.Token{
		AccessToken:  stored.AccessToken,
		RefreshToken: stored.RefreshToken,
		TokenType:    stored.TokenType,
		Expiry:       stored.Expiry,
	}

	// 3. Refresh the token if needed (token source)
	newToken, err := s.refreshAccessTokenIfExpired(ctx, token)
	if err != nil {
		return fmt.Errorf("token refresh error: %w", err)
	}

	// 4. Use the new (or existing) token to fetch data
	purchased, used, remaining, err := s.fetchHoursData(newToken, stored.RealmID)
	if err != nil {
		return fmt.Errorf("fetch hours error: %w", err)
	}

	// 5. Update sub-document in Firestore with new hours + updated token info
	stored.AccessToken = newToken.AccessToken
	stored.RefreshToken = newToken.RefreshToken
	stored.Expiry = newToken.Expiry
	stored.HoursPurchased = purchased
	stored.HoursUsed = used
	stored.HoursRemaining = remaining

	if err := s.storeIntuitData(ctx, parentID, stored); err != nil {
		return fmt.Errorf("failed to update stored token/hours: %w", err)
	}

	return nil
}

// fetchHoursData is where you'd call QuickBooks to get "Hours Purchased" from Invoices
// and combine it with "Hours Used" from your own logic, or from QBO if that’s how you track usage.
// This is a placeholder showing how you'd do a QBO query.
func (s *OAuthService) fetchHoursData(token *oauth2.Token, realmID string) (purchased, used, remaining float64, err error) {
	// 1. Create an HTTP client from the token
	client := s.config.Client(context.Background(), token)

	// 2. Query QuickBooks (sandbox) for Invoices
	// You can refine the query or make multiple calls for Payment or Invoice lines, etc.
	q := url.QueryEscape("SELECT * FROM Invoice")
	requestURL := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/query?query=%s", realmID, q)

	resp, err := client.Get(requestURL)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invoice query error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// 3. Parse the JSON response to get line items
	// For brevity, we won't show the entire QBO invoice JSON structure here.
	// Let's pretend we found 10 hours total purchased from the sample data:
	purchased = 10.0

	// 4. "used" might be tracked in your system, or also in QBO (like a timesheet).
	used = 6.0
	remaining = purchased - used

	return purchased, used, remaining, nil
}

// refreshAccessTokenIfExpired uses the underlying oauth2 TokenSource to refresh if needed.
func (s *OAuthService) refreshAccessTokenIfExpired(ctx context.Context, tok *oauth2.Token) (*oauth2.Token, error) {
	// If the token isn't expired, just return it. Otherwise, refresh.
	if tok.Valid() {
		return tok, nil
	}

	// Create a token source from the config using the current token
	tokenSource := s.config.TokenSource(ctx, tok)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("refresh token error: %w", err)
	}
	return newToken, nil
}

// storeIntuitData stores the IntuitIntegration struct in Firestore as a sub-document in "parents/{parentID}".
func (s *OAuthService) storeIntuitData(ctx context.Context, parentID string, data IntuitIntegration) error {
	_, err := s.firestore.Collection("parents").Doc(parentID).Set(ctx, map[string]interface{}{
		"intuitIntegration": data,
	}, firestore.MergeAll)
	return err
}

// -------------------------------------------------------------------
// Below is an OPTIONAL convenience method if you want to quickly fetch
// the parent's hours from Firestore for your frontend.
// You could call this from some /api/parents/hours?parentID=xxx route.
// -------------------------------------------------------------------
func (s *OAuthService) GetParentHours(ctx context.Context, parentID string) (float64, float64, float64, error) {
	docSnap, err := s.firestore.Collection("parents").Doc(parentID).Get(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get parent doc: %w", err)
	}

	var temp struct {
		IntuitIntegration IntuitIntegration `firestore:"intuitIntegration"`
	}
	if err := docSnap.DataTo(&temp); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse doc data: %w", err)
	}

	return temp.IntuitIntegration.HoursPurchased, temp.IntuitIntegration.HoursUsed, temp.IntuitIntegration.HoursRemaining, nil
}
