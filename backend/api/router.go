package api

import (
	"net/http"
	"social-network/websocket" 
	

	"github.com/gorilla/mux"
)

// SetupRouter configures all the API routes for the application.
func SetupRouter(hub *websocket.Hub) http.Handler {
	// This is cleaner than using a global variable.
	userHandlers := NewUserHandlers(hub)

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// --- Public Routes ---
	apiRouter.HandleFunc("/register", userHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", userHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/logout", userHandlers.LogoutHandler).Methods("POST", "OPTIONS")

	// --- Protected Routes (require a valid session cookie) ---
	apiRouter.HandleFunc("/me", AuthMiddleware(userHandlers.CurrentUserHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/follow/{userId}", AuthMiddleware(userHandlers.SendFollowRequestHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/follow/{userId}/accept", AuthMiddleware(userHandlers.AcceptFollowRequestHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/follow/{userId}/decline", AuthMiddleware(userHandlers.DeclineFollowRequestHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/follow/{userId}", AuthMiddleware(userHandlers.UnfollowHandler)).Methods("DELETE", "OPTIONS")
	apiRouter.HandleFunc("/followers/{userId}", userHandlers.ListFollowersHandler).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/following/{userId}", userHandlers.ListFollowingHandler).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/follow-requests", AuthMiddleware(userHandlers.ListPendingFollowRequestsHandler)).Methods("GET", "OPTIONS")

	// --- Profile Endpoints ---
	apiRouter.HandleFunc("/profile/{userId}", userHandlers.GetProfileHandler).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/profile", AuthMiddleware(userHandlers.UpdateProfileHandler)).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/profile/avatar", AuthMiddleware(userHandlers.UploadAvatarHandler)).Methods("POST", "OPTIONS")

	
	// --- WebSocket Route ---
	// Authentication is handled inside the WebSocket handler itself
	apiRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})
	apiRouter.HandleFunc("/make-private/{userId}", AuthMiddleware(userHandlers.MakeProfilePrivateHandler)).Methods("POST", "OPTIONS")

	// Wrap the router with CORS middleware before returning
	return CORSMiddleware(router)
}