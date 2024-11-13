// backend/middleware/auth_middleware.go

package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const userContextKey = contextKey("user")

func AuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			log.Printf("Authorization header: %s", authHeader)

			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
				return
			}
			tokenStr := parts[1]

			// Parse and validate the token
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Ensure the token's signing method is what we expect
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Get claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Log the user ID from the token
			log.Printf("Authenticated user ID: %s", claims["user_id"])

			// Set user info in context
			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Retrieve user info (claims) from context
func GetUserFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(userContextKey).(jwt.MapClaims)
	return claims, ok
}

// ExtractUserIDFromContext extracts the user ID from the context
func ExtractUserIDFromContext(ctx context.Context) (string, error) {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return "", http.ErrNoCookie // Or appropriate error
	}
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", http.ErrNoCookie
	}
	return userID, nil
}
