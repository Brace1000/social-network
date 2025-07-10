package api

import (
	"context"
	"net/http"

	"social-network/services"
)

// AuthMiddleware checks for a valid session and adds the user to the request context.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle pre-flight CORS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		user, err := services.GetUserFromSession(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user to the request context
		ctx := context.WithValue(r.Context(), services.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
