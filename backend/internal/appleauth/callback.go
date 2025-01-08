package appleauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Log local/UTC time for debugging clock drift
	nowLocal := time.Now()
	nowUTC := time.Now().UTC()
	log.Printf("System time (local): %v, (UTC): %v", nowLocal, nowUTC)

	// Parse the form data to get the code (and potentially the "user" JSON)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Code not found in the request", http.StatusBadRequest)
		return
	}

	// Apple provides "user" JSON ONLY on the very first login attempt
	userInfoStr := r.FormValue("user")
	var name string
	if userInfoStr != "" {
		// userInfo looks like: {"name":{"firstName":"Bob","lastName":"Smith"}, ...}
		var userInfo struct {
			Name struct {
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			} `json:"name"`
		}
		if err := json.Unmarshal([]byte(userInfoStr), &userInfo); err == nil {
			name = fmt.Sprintf("%s %s", userInfo.Name.FirstName, userInfo.Name.LastName)
			log.Printf("Received user name from Apple: %s", name)
		}
	}

	// Generate the client_secret JWT for Apple
	clientSecret, err := a.generateClientSecret()
	if err != nil {
		log.Printf("Failed to generate client secret: %v", err)
		http.Error(w, "Failed to generate client secret", http.StatusInternalServerError)
		return
	}

	// Clone the OAuth2 config and set the runtime client secret
	oauthConfig := *a.OAuthConfig
	oauthConfig.ClientSecret = clientSecret

	// Exchange code for an access token (and hopefully ID token)
	token, err := oauthConfig.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		if rErr, ok := err.(*oauth2.RetrieveError); ok {
			log.Printf("Body: %s", rErr.Body)
		}
		http.Error(w, "Failed to exchange authorization code for token", http.StatusBadRequest)
		return
	}

	log.Printf("Apple OAuth exchange successful. AccessToken=%s", token.AccessToken)

	// Apple should also return an ID token
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found in Apple response", http.StatusBadRequest)
		return
	}

	// Validate and parse the ID token to get claims
	claims, err := a.validateIDToken(idTokenStr)
	if err != nil {
		log.Printf("Failed to validate ID token: %v", err)
		http.Error(w, fmt.Sprintf("Failed to validate ID token: %v", err), http.StatusBadRequest)
		return
	}

	// Extract "sub" (Apple's unique user ID)
	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		http.Error(w, "Invalid user ID in ID token", http.StatusBadRequest)
		return
	}

	// Attempt to retrieve email from the token claims
	email, _ := claims["email"].(string)
	emailVerified, _ := claims["email_verified"].(bool)
	if !emailVerified {
		log.Println("Warning: Apple user email is not verified")
	}

	// If Apple doesn't provide email this time, we can see if we have it stored from a prior login
	firestoreClient := a.FirestoreClient
	if email == "" {
		docRef := firestoreClient.Collection("parents").Doc(userID)
		docSnap, err := docRef.Get(context.Background())
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// This is the user's first login but Apple didn't provide an email => can't proceed
				http.Error(w, "Email not provided by Apple and user does not exist", http.StatusBadRequest)
				return
			}
			http.Error(w, "Error checking user record in Firestore", http.StatusInternalServerError)
			return
		}
		oldData := docSnap.Data()
		storedEmail, ok := oldData["email"].(string)
		if !ok || storedEmail == "" {
			http.Error(w, "User record found but no email stored", http.StatusBadRequest)
			return
		}
		email = storedEmail
	}

	// Apple doesn't provide a picture in the claims
	pictureURL := ""

	// Create your own JWT for the front-end, storing userID and email
	tokenClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // one-week expiry
	}
	myJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedJWT, err := myJWT.SignedString([]byte(a.SecretKey))
	if err != nil {
		log.Printf("Failed to create session JWT token: %v", err)
		http.Error(w, "Failed to create session token", http.StatusInternalServerError)
		return
	}

	// Now store/update the user in Firestore
	docRef := firestoreClient.Collection("parents").Doc(userID)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// New user
			_, err = docRef.Set(context.Background(), map[string]interface{}{
				"user_id":             userID,
				"email":               email,
				"name":                name,
				"picture":             pictureURL,
				"access_token":        token.AccessToken,
				"refresh_token":       token.RefreshToken,
				"expiry":              token.Expiry,
				"associated_students": []interface{}{},
				"created_at":          time.Now(),
			})
			if err != nil {
				log.Printf("Failed to create user doc: %v", err)
				http.Error(w, "Failed to create user document in Firestore", http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Error checking user existence in Firestore: %v", err)
			http.Error(w, "Failed to check user existence in Firestore", http.StatusInternalServerError)
			return
		}
	} else {
		// Existing user, update as needed
		existingData := docSnap.Data()
		updates := []firestore.Update{}
		needsUpdate := false

		if existingData["access_token"] != token.AccessToken {
			updates = append(updates, firestore.Update{Path: "access_token", Value: token.AccessToken})
			needsUpdate = true
		}
		if existingData["refresh_token"] != token.RefreshToken && token.RefreshToken != "" {
			updates = append(updates, firestore.Update{Path: "refresh_token", Value: token.RefreshToken})
			needsUpdate = true
		}
		if existingData["expiry"] != token.Expiry {
			updates = append(updates, firestore.Update{Path: "expiry", Value: token.Expiry})
			needsUpdate = true
		}
		if name != "" && existingData["name"] != name {
			updates = append(updates, firestore.Update{Path: "name", Value: name})
			needsUpdate = true
		}
		if existingData["picture"] != pictureURL {
			updates = append(updates, firestore.Update{Path: "picture", Value: pictureURL})
			needsUpdate = true
		}
		if needsUpdate && len(updates) > 0 {
			_, err := docRef.Update(context.Background(), updates)
			if err != nil {
				log.Printf("Failed to update user doc: %v", err)
				http.Error(w, "Failed to update user document in Firestore", http.StatusInternalServerError)
				return
			}
		}
	}

	log.Printf("OAuthCallbackHandler: userID=%s email=%s name=%s", userID, email, name)
	log.Printf("OAuthCallbackHandler: Redirecting to front-end with session token")

	// Redirect the user to your front-end with your newly created JWT in the fragment
	redirectURL := fmt.Sprintf("%s/auth-redirect#%s", "https://lee-tutoring-webapp.web.app", signedJWT)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// generateClientSecret: signs the JWT Apple needs for token exchange
func (a *App) generateClientSecret() (string, error) {
	teamID := a.Config.APPLE_TEAM_ID
	clientID := a.Config.APPLE_CLIENT_ID
	keyID := a.Config.APPLE_KEY_ID

	// Load the .p8 private key from the env variable
	pemEncodedKey := os.Getenv("APPLE_PRIVATE_KEY")
	if pemEncodedKey == "" {
		return "", fmt.Errorf("APPLE_PRIVATE_KEY environment variable is not set")
	}

	block, _ := pem.Decode([]byte(pemEncodedKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block from APPLE_PRIVATE_KEY")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse PKCS8 private key: %v", err)
	}

	ecdsaPrivateKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not an ECDSA key")
	}

	now := time.Now()
	exp := now.Add(180 * 24 * time.Hour) // Apple recommends up to 6 months

	log.Printf("Generating Apple client secret with claims:")
	log.Printf("  TeamID (iss)=%s", teamID)
	log.Printf("  ClientID (sub)=%s", clientID)
	log.Printf("  KeyID (kid)=%s", keyID)
	log.Printf("  iat=%d  exp=%d", now.Unix(), exp.Unix())

	claims := jwt.MapClaims{
		"iss": teamID,     // Apple Team ID
		"iat": now.Unix(), // Issued at
		"exp": exp.Unix(), // Expires
		"aud": "https://appleid.apple.com",
		"sub": clientID, // Your "Services ID"
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = keyID

	signedStr, err := tok.SignedString(ecdsaPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign client secret: %v", err)
	}
	return signedStr, nil
}

// validateIDToken fetches Apple's public keys and verifies the JWT signature & claims
func (a *App) validateIDToken(idToken string) (jwt.MapClaims, error) {
	keySet, err := jwk.Fetch(context.Background(), "https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple's public keys: %v", err)
	}

	// Parse the token, allowing both ES256 and RS256
	token, err := jwt.Parse(idToken, func(tok *jwt.Token) (interface{}, error) {
		alg := tok.Method.Alg()
		if alg != jwt.SigningMethodES256.Alg() && alg != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", alg)
		}

		kidVal, ok := tok.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid found in token header")
		}

		// Look up matching key
		key, found := keySet.LookupKeyID(kidVal)
		if !found {
			return nil, fmt.Errorf("public key not found for kid: %s", kidVal)
		}
		var raw interface{}
		if err := key.Raw(&raw); err != nil {
			return nil, fmt.Errorf("failed to get raw public key: %v", err)
		}
		return raw, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse ID token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid ID token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional checks: iss, aud, etc.
	if claims["iss"] != "https://appleid.apple.com" {
		return nil, fmt.Errorf("invalid issuer in ID token")
	}
	if claims["aud"] != a.Config.APPLE_CLIENT_ID {
		return nil, fmt.Errorf("invalid audience in ID token")
	}

	return claims, nil
}
