// File: backend/internal/googleauth/callback.go
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
	ctx := context.Background()

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in the request", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for a token
	token, err := a.OAuthConfig.Exchange(ctx, code)
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

	payload, err := idtoken.Validate(ctx, idTokenStr, a.Config.GOOGLE_CLIENT_ID)
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

	// Extract optional fields
	name, _ := payload.Claims["name"].(string)
	pictureURL, _ := payload.Claims["picture"].(string)

	client := a.FirestoreClient
	var (
		role    string
		docRef  *firestore.DocumentRef
		docSnap *firestore.DocumentSnapshot
	)

	// 1) Tutor check by company domain
	if strings.HasSuffix(email, "@leetutoring.com") {
		role = "tutor"
		docRef = client.Collection("tutors").Doc(userID)
		docSnap, err = docRef.Get(ctx)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// Create new tutor
				tutorData := map[string]interface{}{
					"user_id":       userID,
					"email":         email,
					"name":          name,
					"picture":       pictureURL,
					"access_token":  token.AccessToken,
					"refresh_token": token.RefreshToken,
					"expiry":        token.Expiry,
					"userType":      role,
				}
				if _, err := docRef.Set(ctx, tutorData); err != nil {
					log.Printf("Failed to create tutor document: %v", err)
					http.Error(w, "Failed to create tutor in Firestore", http.StatusInternalServerError)
					return
				}
			} else {
				log.Printf("Error fetching tutor document: %v", err)
				http.Error(w, "Failed to check tutor in Firestore", http.StatusInternalServerError)
				return
			}
		} else {
			// Existing tutor: update as needed
			updates := []firestore.Update{}
			data := docSnap.Data()
			if data["access_token"] != token.AccessToken {
				updates = append(updates, firestore.Update{Path: "access_token", Value: token.AccessToken})
			}
			if token.RefreshToken != "" && data["refresh_token"] != token.RefreshToken {
				updates = append(updates, firestore.Update{Path: "refresh_token", Value: token.RefreshToken})
			}
			if exp, ok := data["expiry"].(time.Time); !ok || !exp.Equal(token.Expiry) {
				updates = append(updates, firestore.Update{Path: "expiry", Value: token.Expiry})
			}
			if data["name"] != name {
				updates = append(updates, firestore.Update{Path: "name", Value: name})
			}
			if data["picture"] != pictureURL {
				updates = append(updates, firestore.Update{Path: "picture", Value: pictureURL})
			}
			if data["userType"] != role {
				updates = append(updates, firestore.Update{Path: "userType", Value: role})
			}
			if len(updates) > 0 {
				if _, err := docRef.Update(ctx, updates); err != nil {
					log.Printf("Failed to update tutor document: %v", err)
					http.Error(w, "Failed to update tutor in Firestore", http.StatusInternalServerError)
					return
				}
			}
		}
	} else {
		// 2) Student check against students.personal.student_email
		query := client.Collection("students").Where("personal.student_email", "==", email).Limit(1)
		iter := query.Documents(ctx)
		docs, err := iter.GetAll()
		if err != nil {
			log.Printf("Error querying students: %v", err)
			http.Error(w, "Failed to query students in Firestore", http.StatusInternalServerError)
			return
		}
		if len(docs) > 0 {
			// Existing student: drop data onto top-level document
			role = "student"
			docSnap = docs[0]
			docRef = docSnap.Ref

			updates := []firestore.Update{}
			data := docSnap.Data()
			if data["access_token"] != token.AccessToken {
				updates = append(updates, firestore.Update{Path: "access_token", Value: token.AccessToken})
			}
			if token.RefreshToken != "" && data["refresh_token"] != token.RefreshToken {
				updates = append(updates, firestore.Update{Path: "refresh_token", Value: token.RefreshToken})
			}
			if exp, ok := data["expiry"].(time.Time); !ok || !exp.Equal(token.Expiry) {
				updates = append(updates, firestore.Update{Path: "expiry", Value: token.Expiry})
			}
			if data["name"] != name {
				updates = append(updates, firestore.Update{Path: "name", Value: name})
			}
			if data["picture"] != pictureURL {
				updates = append(updates, firestore.Update{Path: "picture", Value: pictureURL})
			}
			if data["userType"] != role {
				updates = append(updates, firestore.Update{Path: "userType", Value: role})
			}
			if len(updates) > 0 {
				if _, err := docRef.Update(ctx, updates); err != nil {
					log.Printf("Failed to update student document: %v", err)
					http.Error(w, "Failed to update student in Firestore", http.StatusInternalServerError)
					return
				}
			}
		} else {
			// 3) Parent fallback
			role = "parent"
			docRef = client.Collection("parents").Doc(userID)
			docSnap, err = docRef.Get(ctx)
			if err != nil {
				if status.Code(err) == codes.NotFound {
					parentData := map[string]interface{}{
						"user_id":             userID,
						"email":               email,
						"name":                name,
						"picture":             pictureURL,
						"access_token":        token.AccessToken,
						"refresh_token":       token.RefreshToken,
						"expiry":              token.Expiry,
						"associated_students": []interface{}{},
						"userType":            role,
					}
					if _, err := docRef.Set(ctx, parentData); err != nil {
						log.Printf("Failed to create parent document: %v", err)
						http.Error(w, "Failed to create parent in Firestore", http.StatusInternalServerError)
						return
					}
				} else {
					log.Printf("Error fetching parent document: %v", err)
					http.Error(w, "Failed to check parent in Firestore", http.StatusInternalServerError)
					return
				}
			} else {
				updates := []firestore.Update{}
				data := docSnap.Data()
				if data["access_token"] != token.AccessToken {
					updates = append(updates, firestore.Update{Path: "access_token", Value: token.AccessToken})
				}
				if token.RefreshToken != "" && data["refresh_token"] != token.RefreshToken {
					updates = append(updates, firestore.Update{Path: "refresh_token", Value: token.RefreshToken})
				}
				if exp, ok := data["expiry"].(time.Time); !ok || !exp.Equal(token.Expiry) {
					updates = append(updates, firestore.Update{Path: "expiry", Value: token.Expiry})
				}
				if data["name"] != name {
					updates = append(updates, firestore.Update{Path: "name", Value: name})
				}
				if data["picture"] != pictureURL {
					updates = append(updates, firestore.Update{Path: "picture", Value: pictureURL})
				}
				if data["userType"] != role {
					updates = append(updates, firestore.Update{Path: "userType", Value: role})
				}
				if len(updates) > 0 {
					if _, err := docRef.Update(ctx, updates); err != nil {
						log.Printf("Failed to update parent document: %v", err)
						http.Error(w, "Failed to update parent in Firestore", http.StatusInternalServerError)
						return
					}
				}
			}
		}
	}

	// Generate JWT with correct role
	tokenClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedToken, err := jwtToken.SignedString([]byte(a.SecretKey))
	if err != nil {
		log.Printf("Failed to sign JWT token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("OAuthCallbackHandler: User authenticated: %s (%s), role: %s", userID, email, role)
	// Redirect to the React app's auth-redirect handler
	redirectURL := fmt.Sprintf("%s/auth-redirect#%s", "https://lee-tutoring-webapp.web.app", signedToken)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
