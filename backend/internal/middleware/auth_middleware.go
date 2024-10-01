package middleware

import (
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Perform authentication checks here

		// If authenticated, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}
