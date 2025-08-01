package api

import (
	"context"
	"log"
	"net/http"

	"social-network/database"
)

// Define a new type for our context key to avoid collisions.
type contextKey string

const userContextKey = contextKey("userID")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AuthMiddleware: Processing request to %s", r.URL.Path)

		// 1. Attempt to get the cookie from the request
		cookie, err := r.Cookie("social_network_session")
		if err != nil {
			log.Printf("AuthMiddleware: Cookie 'social_network_session' not found, error: %v", err)
			respondWithError(w, http.StatusUnauthorized, "User not authenticated: missing session cookie")
			return
		}

		sessionToken := cookie.Value
		log.Printf("AuthMiddleware: Found session token: %s", sessionToken)

		if sessionToken == "" {
			log.Println("AuthMiddleware: Session token is empty")
			respondWithError(w, http.StatusUnauthorized, "User not authenticated: invalid session token")
			return
		}

		// 3. Validate the session token against the database
		var userID string // Keep as string since it's a UUID
		query := "SELECT user_id FROM sessions WHERE token = ? AND expiry > CURRENT_TIMESTAMP"
		err = database.DB.QueryRow(query, sessionToken).Scan(&userID)
		if err != nil {
			log.Printf("AuthMiddleware: Invalid session token '%s', error: %v", sessionToken, err)
			respondWithError(w, http.StatusUnauthorized, "User not authenticated: invalid or expired session")
			return
		}

		log.Printf("AuthMiddleware: Successfully authenticated user ID: %s", userID)

		// 4. Add the string user ID to the context (don't convert to int!)
		ctx := context.WithValue(r.Context(), "userID", userID) // userID should be a string (UUID)
next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Your CORS middleware is fine and does not need changes.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie") // Add Cookie to allowed headers
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
