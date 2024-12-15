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

	// Parse form data to ensure form fields are read correctly
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

	// Get user parameter (name) if available — only on first login Apple provides name
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

	// Generate the client secret JWT for Apple
	clientSecret, err := a.generateClientSecret()
	if err != nil {
		log.Printf("Failed to generate client secret: %v", err)
		http.Error(w, "Failed to generate client secret", http.StatusInternalServerError)
		return
	}

	// Create a new OAuth2 config with the generated client secret
	oauthConfig := *a.OAuthConfig
	oauthConfig.ClientSecret = clientSecret

	// Exchange the authorization code for tokens
	// Explicitly add grant_type parameter
	token, err := oauthConfig.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
	)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, "Failed to exchange authorization code for token", http.StatusBadRequest)
		return
	}

	// Extract the ID token
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found in token response", http.StatusBadRequest)
		return
	}

	// Validate the ID token and get claims
	claims, err := a.validateIDToken(idTokenStr)
	if err != nil {
		log.Printf("Failed to validate ID token: %v", err)
		http.Error(w, fmt.Sprintf("Failed to validate ID token: %v", err), http.StatusInternalServerError)
		return
	}

	// Get userID (sub) from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in ID token", http.StatusBadRequest)
		return
	}

	// Attempt to get email from claims
	email, emailPresent := claims["email"].(string)

	// If no email in this login attempt, try to fetch from Firestore (assuming a previous login occurred)
	firestoreClient := a.FirestoreClient
	if !emailPresent || email == "" {
		docRef := firestoreClient.Collection("parents").Doc(userID)
		doc, err := docRef.Get(context.Background())
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// This is first login but no email provided by Apple => fail
				http.Error(w, "Email not provided by Apple and user does not exist", http.StatusBadRequest)
				return
			} else {
				http.Error(w, "Error checking user record in Firestore", http.StatusInternalServerError)
				return
			}
		}

		data := doc.Data()
		storedEmail, ok := data["email"].(string)
		if !ok || storedEmail == "" {
			// User has logged in before but no email stored => fail
			http.Error(w, "User record found but no email stored", http.StatusBadRequest)
			return
		}
		email = storedEmail
	}

	emailVerified, _ := claims["email_verified"].(bool)
	if !emailVerified {
		log.Println("Email not verified")
		// Optionally handle unverified emails differently if required
	}

	// Apple doesn't provide a picture
	pictureURL := ""

	// Create your own JWT for the user
	tokenClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedToken, err := jwtToken.SignedString([]byte(a.SecretKey))
	if err != nil {
		log.Printf("Failed to sign JWT token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Check if user exists in Firestore
	docRef := firestoreClient.Collection("parents").Doc(userID)
	doc, err := docRef.Get(context.Background())

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
		// Existing user, update as needed
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
		if data["name"] != name {
			updates = append(updates, firestore.Update{Path: "name", Value: name})
			needsUpdate = true
		}
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

	// Redirect to your app's dashboard
	redirectURL := fmt.Sprintf("%s/auth-redirect#%s", "https://lee-tutoring-webapp.web.app", signedToken)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// Helper functions
func (a *App) generateClientSecret() (string, error) {
	teamID := a.Config.APPLE_TEAM_ID
	clientID := a.Config.APPLE_CLIENT_ID
	keyID := a.Config.APPLE_KEY_ID

	// Retrieve the raw PEM-encoded private key from the environment variable
	pemEncodedKey := os.Getenv("APPLE_PRIVATE_KEY")
	if pemEncodedKey == "" {
		return "", fmt.Errorf("APPLE_PRIVATE_KEY environment variable is not set")
	}

	// Decode the PEM block
	block, _ := pem.Decode([]byte(pemEncodedKey))
	if block == nil {
		log.Fatalf("Failed to parse PEM block from APPLE_PRIVATE_KEY")
	}

	// Parse the PKCS8 private key
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Printf("ParsePKCS8PrivateKey error: %v\nPEM Block Type: %s\n", err, block.Type)
		return "", fmt.Errorf("failed to parse PKCS8 private key: %v", err)
	}

	// Assert that the key is of type *ecdsa.PrivateKey
	ecdsaPrivateKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		log.Printf("Private key type assertion failed; got type %T\n", key)
		return "", fmt.Errorf("private key is not an ECDSA key")
	}

	// Define the JWT claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
		"exp": now.Add(180 * 24 * time.Hour).Unix(), // Set to 6 months as per Apple's guidelines
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	// Sign the token with the ECDSA private key
	clientSecret, err := token.SignedString(ecdsaPrivateKey)
	if err != nil {
		log.Printf("Signing error: %v\n", err)
		return "", fmt.Errorf("failed to sign client secret: %v", err)
	}

	return clientSecret, nil
}

func (a *App) validateIDToken(idToken string) (jwt.MapClaims, error) {
	keySet, err := jwk.Fetch(context.Background(), "https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple's public keys: %v", err)
	}

	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid found in token header")
		}

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

	if claims["aud"] != a.Config.APPLE_CLIENT_ID {
		return nil, fmt.Errorf("invalid audience in ID token")
	}
	if claims["iss"] != "https://appleid.apple.com" {
		return nil, fmt.Errorf("invalid issuer in ID token")
	}

	return claims, nil
}
