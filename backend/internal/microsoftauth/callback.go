// backend/internal/microsoftauth/callback.go

package microsoftauth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/coreos/go-oidc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	// Exchange the authorization code for a token
	token, err := a.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract the ID Token from OAuth2 token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found", http.StatusBadRequest)
		return
	}

	// Manually fetch the OpenID Connect discovery document
	providerURL := "https://login.microsoftonline.com/common/v2.0"
	resp, err := http.Get(providerURL + "/.well-known/openid-configuration")
	if err != nil {
		http.Error(w, "Failed to get provider configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get provider configuration: "+resp.Status, http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read provider configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var providerData struct {
		Issuer  string `json:"issuer"`
		JWKSURI string `json:"jwks_uri"`
	}

	if err := json.Unmarshal(body, &providerData); err != nil {
		http.Error(w, "Failed to parse provider configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set up an OpenID Connect verifier with the dynamic issuer
	oidcConfig := &oidc.Config{
		ClientID:             a.Config.MICROSOFT_CLIENT_ID,
		SkipIssuerCheck:      true, // We'll handle issuer check manually
		SupportedSigningAlgs: []string{"RS256"},
	}

	keySet := oidc.NewRemoteKeySet(context.Background(), providerData.JWKSURI)
	verifier := oidc.NewVerifier(providerData.Issuer, keySet, oidcConfig)

	// Verify the ID token
	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Manually validate the issuer
	expectedIssuerPrefix := "https://login.microsoftonline.com/"
	expectedIssuerSuffix := "/v2.0"
	issuer := idToken.Issuer

	if !strings.HasPrefix(issuer, expectedIssuerPrefix) || !strings.HasSuffix(issuer, expectedIssuerSuffix) {
		http.Error(w, "Invalid issuer: "+issuer, http.StatusUnauthorized)
		return
	}

	// Extract user claims
	var claims struct {
		Sub               string `json:"sub"`
		PreferredUsername string `json:"preferred_username"`
		Email             string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID := claims.Sub
	email := claims.Email
	if email == "" {
		email = claims.PreferredUsername
	}

	// Store userID and email in session
	session, err := a.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = userID
	session.Values["user_email"] = email // Store the email in the session

	// Save the session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
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
				http.Error(w, "Failed to create user document in Firestore: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Redirect to the parent dashboard (automatic association will occur there)
			http.Redirect(w, r, "/parentdashboard", http.StatusSeeOther)
			return
		} else {
			http.Error(w, "Failed to check user existence in Firestore: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

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
			http.Error(w, "Failed to update user tokens in Firestore: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "https://lee-tutoring-webapp.web.app/parentdashboard", http.StatusSeeOther)

}
