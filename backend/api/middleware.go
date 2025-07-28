package api

import (
	"context"
	"log"
	"net/http"

	"social-network/database"
)

// AuthMiddleware checks for a valid session and adds the user to the request context.

// Define a new type for our context key to avoid collisions.
type contextKey string

const userContextKey = contextKey("userID")

// AuthMiddleware is a middleware that verifies a user's session cookie.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Attempt to get the cookie from the request.
		cookie, err := r.Cookie("social_network_session")
		if err != nil {
			// This is the most common failure point. If the cookie is not found,
			// http.ErrNoCookie is returned.
			log.Printf("AuthMiddleware: Cookie not found, error: %v", err)
			respondWithError(w, http.StatusUnauthorized, "User not authenticated: missing session cookie")
			return
		}

		// 2. Get the session token (value) from the cookie.
		sessionToken := cookie.Value
		if sessionToken == "" {
			log.Println("AuthMiddleware: Session token is empty")
			respondWithError(w, http.StatusUnauthorized, "User not authenticated: invalid session token")
			return
		}

		// 3. Validate the session token against the database.
		var userID int
		// This SQL query finds the user_id associated with the session token
		// and also checks that the session has not expired.
		query := "SELECT user_id FROM sessions WHERE token = ? AND expiry > CURRENT_TIMESTAMP"
		err = database.DB.QueryRow(query, sessionToken).Scan(&userID)
		if err != nil {
			// If QueryRow returns an error (like sql.ErrNoRows), the session is
			// invalid, expired, or doesn't exist.
			log.Printf("AuthMiddleware: Invalid session token '%s', error: %v", sessionToken, err)
			respondWithError(w, http.StatusUnauthorized, "User not authenticated: invalid or expired session")
			return
		}

		// 4. If we get here, the user is authenticated!
		// Add the user ID to the request context so subsequent handlers can access it.
		ctx := context.WithValue(r.Context(), userContextKey, userID)

		// Call the next handler in the chain, passing the new context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS middleware for development
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
