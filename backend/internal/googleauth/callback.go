// backend/internal/googleauth/callback.go

package googleauth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get user information from the ID token
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found in token response", http.StatusBadRequest)
		return
	}

	payload, err := idtoken.Validate(context.Background(), idTokenStr, a.Config.GOOGLE_CLIENT_ID)
	if err != nil {
		log.Printf("Failed to validate ID token: %v", err)
		http.Error(w, "Failed to validate ID token", http.StatusInternalServerError)
		return
	}

	userID, ok := payload.Claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in ID token", http.StatusBadRequest)
		return
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		http.Error(w, "Invalid email in ID token", http.StatusBadRequest)
		return
	}

	// Extract name
	name, ok := payload.Claims["name"].(string)
	if !ok {
		log.Println("Name not found in ID token")
		name = ""
	}

	// Extract picture URL
	pictureURL, ok := payload.Claims["picture"].(string)
	if !ok {
		log.Println("Picture URL not found in ID token")
		pictureURL = ""
	}

	// Generate a JWT token with role (and without associated_students for tutors)
	tokenClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // Token expires in 7 days
	}

	// Determine the Firestore collection based on the email domain and set the role
	collectionName := "parents"
	if strings.HasSuffix(email, "@leetutoring.com") {
		collectionName = "tutors"
		tokenClaims["role"] = "tutor"
	} else {
		tokenClaims["role"] = "parent"
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	// Use the secret key to sign the token
	signedToken, err := jwtToken.SignedString([]byte(a.SecretKey))
	if err != nil {
		log.Printf("Failed to sign JWT token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Reference to the user's document in the chosen collection
	docRef := firestoreClient.Collection(collectionName).Doc(userID)
	doc, err := docRef.Get(context.Background())

	if err != nil {
		if status.Code(err) == codes.NotFound {
			// User doesn't exist, create a new document.
			// For tutors, we do not create an "associated_students" field.
			var userData map[string]interface{}
			if collectionName == "tutors" {
				userData = map[string]interface{}{
					"user_id":       userID,
					"email":         email,
					"name":          name,
					"picture":       pictureURL,
					"access_token":  token.AccessToken,
					"refresh_token": token.RefreshToken,
					"expiry":        token.Expiry,
				}
			} else {
				userData = map[string]interface{}{
					"user_id":             userID,
					"email":               email,
					"name":                name,
					"picture":             pictureURL,
					"access_token":        token.AccessToken,
					"refresh_token":       token.RefreshToken,
					"expiry":              token.Expiry,
					"associated_students": []interface{}{},
				}
			}
			_, err = docRef.Set(context.Background(), userData)
			if err != nil {
				http.Error(w, "Failed to create user document in Firestore", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to check user existence in Firestore", http.StatusInternalServerError)
			return
		}
	} else {
		// User exists, optionally update tokens if they have changed.
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
		// Update name if it has changed.
		if data["name"] != name {
			updates = append(updates, firestore.Update{Path: "name", Value: name})
			needsUpdate = true
		}
		// Update picture URL if it has changed.
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

	// Redirect to the React app's dashboard route with the token in the URL fragment.
	redirectURL := fmt.Sprintf("%s/auth-redirect#%s", "https://lee-tutoring-webapp.web.app", signedToken)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
