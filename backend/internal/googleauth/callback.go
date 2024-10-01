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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use the existing Firestore client from App
	firestoreClient := a.FirestoreClient

	// Get user information from the ID token
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found", http.StatusBadRequest)
		return
	}

	payload, err := idtoken.Validate(context.Background(), idTokenStr, a.Config.GOOGLE_CLIENT_ID)
	if err != nil {
		http.Error(w, "Failed to validate ID token", http.StatusInternalServerError)
		return
	}

	userID, ok := payload.Claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	// Store userID in session
	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = userID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

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

			// Redirect to the parent intake page
			http.Redirect(w, r, "/parentintake", http.StatusSeeOther)
			return
		} else {
			http.Error(w, "Failed to check user existence in Firestore", http.StatusInternalServerError)
			return
		}
	}

	// User exists, optionally update tokens if they have changed
	data := doc.Data()
	if data["access_token"] != token.AccessToken || data["refresh_token"] != token.RefreshToken {
		_, err = docRef.Update(context.Background(), []firestore.Update{
			{Path: "access_token", Value: token.AccessToken},
			{Path: "refresh_token", Value: token.RefreshToken},
			{Path: "expiry", Value: token.Expiry},
		})
		if err != nil {
			http.Error(w, "Failed to update user tokens in Firestore", http.StatusInternalServerError)
			return
		}
	}

	// Check if associated_students array exists and is not empty
	associatedStudents, ok := data["associated_students"].([]interface{})
	if !ok || len(associatedStudents) == 0 {
		// Initialize the associated_students field as an empty array if it doesn't exist or is empty
		associatedStudents = []interface{}{}
		_, err = docRef.Update(context.Background(), []firestore.Update{
			{Path: "associated_students", Value: associatedStudents},
		})
		if err != nil {
			http.Error(w, "Failed to initialize associated_students in Firestore", http.StatusInternalServerError)
			return
		}

		// Redirect to the parent intake page
		http.Redirect(w, r, "/parentintake", http.StatusSeeOther)
		return
	}

	// Redirect to the parent dashboard
	http.Redirect(w, r, "/parentdashboard", http.StatusSeeOther)
}
