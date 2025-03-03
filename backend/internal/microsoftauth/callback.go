// backend/internal/microsoftauth/callback.go

package microsoftauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/v4"
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

	// Extract the ID Token from OAuth2 token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found", http.StatusBadRequest)
		return
	}

	// Fetch the OpenID Connect discovery document
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

	// Extract user claims from ID token
	var claims struct {
		Sub               string `json:"sub"`
		PreferredUsername string `json:"preferred_username"`
		Email             string `json:"email"`
		Name              string `json:"name"`
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

	name := claims.Name
	if name == "" {
		// Fetch user profile from Microsoft Graph API
		req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
		if err != nil {
			http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			http.Error(w, "Failed to get user profile: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Failed to get user profile: "+resp.Status, http.StatusInternalServerError)
			return
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read user profile response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var userData struct {
			ID                string `json:"id"`
			DisplayName       string `json:"displayName"`
			Mail              string `json:"mail"`
			UserPrincipalName string `json:"userPrincipalName"`
		}
		if err := json.Unmarshal(body, &userData); err != nil {
			http.Error(w, "Failed to parse user profile: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update userID, name, and email based on userData
		userID = userData.ID
		name = userData.DisplayName
		if name == "" {
			name = userData.UserPrincipalName
		}
		if email == "" {
			email = userData.Mail
			if email == "" {
				email = userData.UserPrincipalName
			}
		}
	}

	// Since we cannot get a direct URL to the user's profile picture, we can store a flag indicating the user has a profile picture
	// Alternatively, you can set pictureURL to an endpoint in your backend that serves the picture when requested
	hasProfilePicture := false

	// Check if the user has a profile picture
	photoReq, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/photo", nil)
	if err == nil {
		photoReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
		photoResp, err := http.DefaultClient.Do(photoReq)
		if err == nil && photoResp.StatusCode == http.StatusOK {
			hasProfilePicture = true
		}
		if photoResp != nil {
			photoResp.Body.Close()
		}
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
				"has_profile_picture": hasProfilePicture,
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
		// Update has_profile_picture if it has changed
		if data["has_profile_picture"] != hasProfilePicture {
			updates = append(updates, firestore.Update{Path: "has_profile_picture", Value: hasProfilePicture})
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
