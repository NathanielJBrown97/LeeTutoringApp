// backend/internal/googleauth/callback.go

package googleauth

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	// Exchange the authorization code for a token
	token, err := a.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
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

	// Store userID and email in session
	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = userID
	session.Values["user_email"] = email // Store the email in the session

	// Save the session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Check if the user already exists in Firestore
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
		if data["refresh_token"] != token.RefreshToken {
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

	// Redirect to the React app's dashboard route
	http.Redirect(w, r, "http://localhost:3000/parentdashboard", http.StatusSeeOther)
}
