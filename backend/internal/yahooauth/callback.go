// backend/internal/yahooauth/callback.go

package yahooauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("OAuthCallbackHandler triggered")

	oauthStateCookie, err := r.Cookie("oauthstate")
	if err != nil {
		log.Printf("Error retrieving oauthstate cookie: %v", err)
		http.Error(w, "No OAuth state cookie", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state != oauthStateCookie.Value {
		log.Printf("Invalid OAuth state: expected %s, got %s", oauthStateCookie.Value, state)
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	// Check if there's an error in the query parameters
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errorDescription := r.URL.Query().Get("error_description")
		log.Printf("OAuth error: %s - %s", errParam, errorDescription)
		http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errParam, errorDescription), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in the request", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for a token
	token, err := a.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, "Failed to exchange authorization code for token", http.StatusBadRequest)
		return
	}

	// Use the access token to get user information from Yahoo
	client := a.OAuthConfig.Client(context.Background(), token)
	userInfoURL := "https://api.login.yahoo.com/openid/v1/userinfo"

	resp, err := client.Get(userInfoURL)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to get user info: %s", body)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	var userInfo struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Locale        string `json:"locale"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Failed to parse user info: %v", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	userID := userInfo.Sub
	email := userInfo.Email
	name := userInfo.Name
	pictureURL := userInfo.Picture

	// Generate a JWT token
	tokenClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // Token expires in 7 days
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	// Use the secret key
	signedToken, err := jwtToken.SignedString([]byte(a.SecretKey))
	if err != nil {
		log.Printf("Failed to sign JWT token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Determine the Firestore collection based on the email domain
	collectionName := "parents"
	if strings.HasSuffix(email, "@leetutoring.com") {
		collectionName = "tutors"
	}

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Reference to the user's document in the chosen collection
	docRef := firestoreClient.Collection(collectionName).Doc(userID)
	doc, err := docRef.Get(context.Background())

	if err != nil {
		if status.Code(err) == codes.NotFound {
			// User doesn't exist, create a new document
			_, err = docRef.Set(context.Background(), map[string]interface{}{
				"user_id":             userID,
				"email":               email,
				"name":                name,
				"picture":             pictureURL,
				"access_token":        token.AccessToken,
				"refresh_token":       token.RefreshToken,
				"expiry":              token.Expiry,
				"associated_students": []interface{}{},
			})
			if err != nil {
				http.Error(w, "Failed to create user document in Firestore", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to check user existence in Firestore", http.StatusInternalServerError)
			return
		}
	} else {
		// User exists, optionally update tokens if they have changed
		data := doc.Data()
		needsUpdate := false
		updates := []firestore.Update{}

		if data["access_token"] != token.AccessToken {
			updates = append(updates, firestore.Update{Path: "access_token", Value: token.AccessToken})
			needsUpdate = true
		}
		if data["refresh_token"] != token.RefreshToken && token.RefreshToken != "" {
			updates = append(updates, firestore.Update{Path: "refresh_token", Value: token.RefreshToken})
			needsUpdate = true
		}
		if data["expiry"] != token.Expiry {
			updates = append(updates, firestore.Update{Path: "expiry", Value: token.Expiry})
			needsUpdate = true
		}
		// Update name if it has changed
		if data["name"] != name {
			updates = append(updates, firestore.Update{Path: "name", Value: name})
			needsUpdate = true
		}
		// Update picture URL if it has changed
		if data["picture"] != pictureURL {
			updates = append(updates, firestore.Update{Path: "picture", Value: pictureURL})
			needsUpdate = true
		}

		if needsUpdate && len(updates) > 0 {
			_, err = docRef.Update(context.Background(), updates)
			if err != nil {
				http.Error(w, "Failed to update user tokens in Firestore", http.StatusInternalServerError)
				return
			}
		}
	}

	log.Printf("OAuthCallbackHandler: User authenticated: %s (%s)", userID, email)
	log.Printf("OAuthCallbackHandler: Redirecting to dashboard")

	// Redirect to the React app's dashboard route with the token in the URL fragment
	redirectURL := fmt.Sprintf("%s/auth-redirect#%s", "https://lee-tutoring-webapp.web.app", signedToken)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
