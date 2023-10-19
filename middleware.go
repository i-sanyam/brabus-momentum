package main

import (
	"net/http"
	"os"
)

func authMiddleware(tokenType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the request has a valid authentication token
			authToken := r.Header.Get("Authorization")
			var validToken string
			if tokenType == "read" {
				validToken = os.Getenv("READ_AUTH_TOKEN")
			} else if tokenType == "write" {
				validToken = os.Getenv("WRITE_AUTH_TOKEN")
			} else {
				http.Error(w, "Invalid token type", http.StatusInternalServerError)
				return
			}
			if authToken != validToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
