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
	chatHandlers := NewChatHandlers(hub) // Add chat handlers

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	auth := apiRouter.PathPrefix("").Subrouter()
	auth.Use(AuthMiddleware)

	// --- Public Routes ---
	apiRouter.HandleFunc("/register", userHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", userHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/logout", userHandlers.LogoutHandler).Methods("POST", "OPTIONS")

	// Authentication is handled inside the WebSocket handler itself
	apiRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// --- User & Follower Routes ---
	apiRouter.Handle("/me", AuthMiddleware(http.HandlerFunc(userHandlers.CurrentUserHandler))).Methods("GET", "OPTIONS")
	apiRouter.Handle("/users", AuthMiddleware(http.HandlerFunc(userHandlers.GetAllUsersHandler))).Methods("GET", "OPTIONS")
	apiRouter.Handle("/follow/{userId}", AuthMiddleware(http.HandlerFunc(userHandlers.FollowRequestHandler))).Methods("POST", "OPTIONS")
	apiRouter.Handle("/my-follow-requests", AuthMiddleware(http.HandlerFunc(userHandlers.GetMyFollowRequestsHandler))).Methods("GET", "OPTIONS")
	apiRouter.Handle("/follow-requests/{requestId}/accept", AuthMiddleware(http.HandlerFunc(userHandlers.AcceptFollowRequestHandler))).Methods("POST", "OPTIONS")
	apiRouter.Handle("/follow-requests/{requestId}/decline", AuthMiddleware(http.HandlerFunc(userHandlers.DeclineFollowRequestHandler))).Methods("POST", "OPTIONS")
	apiRouter.Handle("/follow-requests/{requestId}/cancel", AuthMiddleware(http.HandlerFunc(userHandlers.CancelFollowRequestHandler))).Methods("POST", "OPTIONS")
	apiRouter.Handle("/follow-requests", AuthMiddleware(http.HandlerFunc(userHandlers.ListPendingFollowRequestsHandler))).Methods("GET", "OPTIONS")

	// --- Notification Routes ---
	apiRouter.Handle("/notifications", AuthMiddleware(http.HandlerFunc(userHandlers.GetNotificationsHandler))).Methods("GET", "OPTIONS")
	apiRouter.Handle("/notifications/{notificationId}/read", AuthMiddleware(http.HandlerFunc(userHandlers.MarkNotificationAsReadHandler))).Methods("POST", "OPTIONS")
	apiRouter.Handle("/unfollow/{userId}", AuthMiddleware(http.HandlerFunc(userHandlers.UnfollowHandler))).Methods("POST", "OPTIONS") // Using POST for consistency
	apiRouter.Handle("/followers/{userId}", AuthMiddleware(http.HandlerFunc(userHandlers.ListFollowersHandler))).Methods("GET", "OPTIONS")
	apiRouter.Handle("/following/{userId}", AuthMiddleware(http.HandlerFunc(userHandlers.ListFollowingHandler))).Methods("GET", "OPTIONS")
	apiRouter.Handle("/follow-status/{userId}", AuthMiddleware(http.HandlerFunc(userHandlers.CheckFollowRequestStatusHandler))).Methods("GET", "OPTIONS")

	// --- Profile Routes ---
	apiRouter.Handle("/profile/{userId}", AuthMiddleware(http.HandlerFunc(userHandlers.GetProfileHandler))).Methods("GET", "OPTIONS")
	apiRouter.Handle("/profile", AuthMiddleware(http.HandlerFunc(userHandlers.UpdateProfileHandler))).Methods("PUT", "OPTIONS")
	apiRouter.Handle("/profile/avatar", AuthMiddleware(http.HandlerFunc(userHandlers.UploadAvatarHandler))).Methods("POST", "OPTIONS")
	apiRouter.Handle("/profile/make-private", AuthMiddleware(http.HandlerFunc(userHandlers.MakeProfilePrivateHandler))).Methods("POST", "OPTIONS")
	apiRouter.Handle("/profile/toggle-privacy", AuthMiddleware(http.HandlerFunc(userHandlers.ToggleProfilePrivacyHandler))).Methods("POST", "OPTIONS")

	// --- Post & Feed Routes ---
	auth.HandleFunc("/posts", postHandlers.CreatePostHandler).Methods("POST")
	auth.HandleFunc("/posts/feed", postHandlers.GetFeedPostsHandler).Methods("GET")
	auth.HandleFunc("/posts/{postID}/comment", postHandlers.CreateCommentHandler).Methods("POST")
	auth.HandleFunc("/posts/{postID}/like", postHandlers.LikePostHandler).Methods("POST")
	auth.HandleFunc("/comments/{commentID}/like", postHandlers.LikeCommentHandler).Methods("POST")
	// Gets the message history with a specific user.
	auth.HandleFunc("/chats/private/{userID}", chatHandlers.GetPrivateConversationHandler).Methods("GET")

	// Gets the message history for a specific group.
	auth.HandleFunc("/chats/group/{groupID}", chatHandlers.GetGroupConversationHandler).Methods("GET")

	// Wrap the router with CORS middleware before returning
	return CORSMiddleware(router)
}
