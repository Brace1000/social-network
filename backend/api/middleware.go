package api

import (
	"context"
	"net/http"

	"social-network/database"
)

// AuthMiddleware checks for a valid session and adds the user to the request context.

type contextKey string

const userContextKey = contextKey("userID")

// AuthMiddleware is a middleware that verifies a user's session cookie.
// ...
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the session cookie from the request
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized: missing session cookie", http.StatusUnauthorized)
			return
		}
		sessionToken := cookie.Value
		// ...

		// 3. Validate the session token against the database.
		// THE FIX IS HERE:
		var userID string
		query := "SELECT user_id FROM sessions WHERE token = ? AND expiry > CURRENT_TIMESTAMP"
		err = database.DB.QueryRow(query, sessionToken).Scan(&userID) 
		if err != nil {
		
			return
		}

		// 4. Add the string user ID to the context.
		ctx := context.WithValue(r.Context(), userContextKey, userID)
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
