// backend/internal/googleauth/callback.go

package googleauth

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	// Generate a JWT token without associated_students
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

		if needsUpdate {
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
