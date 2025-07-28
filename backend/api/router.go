package api

import (
	"net/http"

	"social-network/websocket"

	"github.com/gorilla/mux"
)

// SetupRouter configures all the API routes for the application.
func SetupRouter(hub *websocket.Hub) http.Handler {
	// Instantiate all handler groups
	userHandlers := NewUserHandlers(hub)
	postHandlers := NewPostHandlers()

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// --- Public Routes ---
	apiRouter.HandleFunc("/register", userHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", userHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/logout", userHandlers.LogoutHandler).Methods("POST", "OPTIONS")
	
	// Authentication is handled inside the WebSocket handler itself
	apiRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// --- User & Follower Routes ---
	apiRouter.HandleFunc("/me",  AuthMiddleware(userHandlers.CurrentUserHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users", AuthMiddleware(userHandlers.GetAllUsersHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/follow/{userId}", AuthMiddleware(userHandlers.FollowRequestHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/my-follow-requests", AuthMiddleware(userHandlers.GetMyFollowRequestsHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/follow-requests/{requestId}/accept", AuthMiddleware(userHandlers.AcceptFollowRequestHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/follow-requests/{requestId}/decline", AuthMiddleware(userHandlers.DeclineFollowRequestHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/follow-requests/{requestId}/cancel", AuthMiddleware(userHandlers.CancelFollowRequestHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/follow-requests", AuthMiddleware(userHandlers.ListPendingFollowRequestsHandler)).Methods("GET", "OPTIONS")
	
	// --- Notification Routes ---
	apiRouter.HandleFunc("/notifications", AuthMiddleware(userHandlers.GetNotificationsHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/notifications/{notificationId}/read", AuthMiddleware(userHandlers.MarkNotificationAsReadHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/unfollow/{userId}", AuthMiddleware(userHandlers.UnfollowHandler)).Methods("POST", "OPTIONS") // Using POST for consistency
	apiRouter.HandleFunc("/followers/{userId}", AuthMiddleware(userHandlers.ListFollowersHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/following/{userId}", AuthMiddleware(userHandlers.ListFollowingHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/follow-status/{userId}", AuthMiddleware(userHandlers.CheckFollowRequestStatusHandler)).Methods("GET", "OPTIONS")

	// --- Profile Routes ---
	apiRouter.HandleFunc("/profile/{userId}", AuthMiddleware(userHandlers.GetProfileHandler)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/profile", AuthMiddleware(userHandlers.UpdateProfileHandler)).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/profile/avatar", AuthMiddleware(userHandlers.UploadAvatarHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/profile/make-private", AuthMiddleware(userHandlers.MakeProfilePrivateHandler)).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/profile/toggle-privacy", AuthMiddleware(userHandlers.ToggleProfilePrivacyHandler)).Methods("POST", "OPTIONS")

	// --- Post & Feed Routes ---
	apiRouter.HandleFunc("/posts", postHandlers.CreatePostHandler).Methods("POST")
	apiRouter.HandleFunc("/posts/feed", postHandlers.GetFeedPostsHandler).Methods("GET")
	apiRouter.HandleFunc("/posts/{postID}/comment", postHandlers.CreateCommentHandler).Methods("POST")

	// Wrap the router with CORS middleware before returning
	return CORSMiddleware(router)
}
