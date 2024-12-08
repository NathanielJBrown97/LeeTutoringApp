// backend/internal/appleauth/callback.go

package appleauth

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Code not found in the request", http.StatusBadRequest)
		return
	}

	// Get 'user' parameter if available
	userInfoStr := r.FormValue("user")
	var name string
	if userInfoStr != "" {
		var userInfo struct {
			Name struct {
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			} `json:"name"`
		}
		err := json.Unmarshal([]byte(userInfoStr), &userInfo)
		if err == nil {
			name = fmt.Sprintf("%s %s", userInfo.Name.FirstName, userInfo.Name.LastName)
		}
	}

	// Generate the client secret JWT
	clientSecret, err := a.generateClientSecret()
	if err != nil {
		log.Printf("Failed to generate client secret: %v", err)
		http.Error(w, "Failed to generate client secret", http.StatusInternalServerError)
		return
	}

	// Create a new OAuth2 config with the client secret
	oauthConfig := *a.OAuthConfig
	oauthConfig.ClientSecret = clientSecret

	// Exchange the authorization code for tokens
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, "Failed to exchange authorization code for token", http.StatusBadRequest)
		return
	}

	// Extract the ID token from the token response
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found in token response", http.StatusBadRequest)
		return
	}

	// Validate the ID token and extract claims
	claims, err := a.validateIDToken(idTokenStr)
	if err != nil {
		log.Printf("Failed to validate ID token: %v", err)
		http.Error(w, "Failed to validate ID token", http.StatusInternalServerError)
		return
	}

	// Extract user information from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in ID token", http.StatusBadRequest)
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, "Invalid email in ID token", http.StatusBadRequest)
		return
	}

	emailVerified, _ := claims["email_verified"].(bool)
	if !emailVerified {
		log.Println("Email not verified")
	}

	// Note: Apple does not provide a picture URL
	pictureURL := ""

	// Generate a JWT token for your app
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

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Reference to the parent's document
	docRef := firestoreClient.Collection("parents").Doc(userID)
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

	log.Printf("OAuthCallbackHandler: Received code=%s", code)
	log.Printf("OAuthCallbackHandler: User authenticated: %s (%s)", userID, email)
	log.Printf("OAuthCallbackHandler: Redirecting to dashboard")

	// Redirect to the React app's dashboard route with the token in the URL fragment
	redirectURL := fmt.Sprintf("%s/auth-redirect#%s", "https://lee-tutoring-webapp.web.app", signedToken)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// Helper functions

func (a *App) generateClientSecret() (string, error) {
	privateKeyData := a.Config.APPLE_PRIVATE_KEY
	teamID := a.Config.APPLE_TEAM_ID
	clientID := a.Config.APPLE_CLIENT_ID
	keyID := a.Config.APPLE_KEY_ID

	block, _ := pem.Decode([]byte(privateKeyData))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}

	ecdsaPrivateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse EC private key: %v", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	clientSecret, err := token.SignedString(ecdsaPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign client secret: %v", err)
	}

	return clientSecret, nil
}

func (a *App) validateIDToken(idToken string) (jwt.MapClaims, error) {
	// Fetch Apple's public keys
	keySet, err := jwk.Fetch(context.Background(), "https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple's public keys: %v", err)
	}

	// Parse the ID token without verification to extract the header
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header (kid)
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid found in token header")
		}

		// Find the corresponding public key in Apple's keys
		key, found := keySet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("public key not found for kid: %s", kid)
		}

		var pubKey interface{}
		if err := key.Raw(&pubKey); err != nil {
			return nil, fmt.Errorf("failed to get public key: %v", err)
		}
		return pubKey, nil
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

	// Verify audience and issuer
	if claims["aud"] != a.Config.APPLE_CLIENT_ID {
		return nil, fmt.Errorf("invalid audience in ID token")
	}
	if claims["iss"] != "https://appleid.apple.com" {
		return nil, fmt.Errorf("invalid issuer in ID token")
	}

	return claims, nil
}
