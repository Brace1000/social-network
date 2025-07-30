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
	chatHandlers := NewChatHandlers(hub)

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// --- Public Routes ---
	apiRouter.HandleFunc("/register", userHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", userHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/logout", userHandlers.LogoutHandler).Methods("POST", "OPTIONS")

	// --- WebSocket Route ---
	// Authentication is handled inside the WebSocket handler itself
	apiRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// --- Protected Routes Group ---
	// All routes that need authentication will be attached to this `auth` sub-router.
	auth := apiRouter.PathPrefix("").Subrouter()
	auth.Use(AuthMiddleware)  

	// --- User & Follower Routes ---
	auth.HandleFunc("/me", userHandlers.CurrentUserHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/users", userHandlers.GetAllUsersHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/follow/{userId}", userHandlers.FollowRequestHandler).Methods("POST", "OPTIONS")
	auth.HandleFunc("/my-follow-requests", userHandlers.GetMyFollowRequestsHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/follow-requests/{requestId}/accept", userHandlers.AcceptFollowRequestHandler).Methods("POST", "OPTIONS")
	auth.HandleFunc("/follow-requests/{requestId}/decline", userHandlers.DeclineFollowRequestHandler).Methods("POST", "OPTIONS")
	auth.HandleFunc("/follow-requests/{requestId}/cancel", userHandlers.CancelFollowRequestHandler).Methods("POST", "OPTIONS")
	auth.HandleFunc("/follow-requests", userHandlers.ListPendingFollowRequestsHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/unfollow/{userId}", userHandlers.UnfollowHandler).Methods("POST", "OPTIONS")
	auth.HandleFunc("/followers/{userId}", userHandlers.ListFollowersHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/following/{userId}", userHandlers.ListFollowingHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/follow-status/{userId}", userHandlers.CheckFollowRequestStatusHandler).Methods("GET", "OPTIONS")

	// --- Notification Routes ---
	auth.HandleFunc("/notifications", userHandlers.GetNotificationsHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/notifications/{notificationId}/read", userHandlers.MarkNotificationAsReadHandler).Methods("POST", "OPTIONS")
	
	// --- Profile Routes ---
	auth.HandleFunc("/profile/{userId}", userHandlers.GetProfileHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/profile", userHandlers.UpdateProfileHandler).Methods("PUT", "OPTIONS")
	auth.HandleFunc("/profile/avatar", userHandlers.UploadAvatarHandler).Methods("POST", "OPTIONS")
	auth.HandleFunc("/profile/toggle-privacy", userHandlers.ToggleProfilePrivacyHandler).Methods("POST", "OPTIONS")

	// --- Post & Feed Routes ---
	auth.HandleFunc("/posts", postHandlers.CreatePostHandler).Methods("POST")
	auth.HandleFunc("/posts/feed", postHandlers.GetFeedPostsHandler).Methods("GET")
	auth.HandleFunc("/posts/{postID}/comment", postHandlers.CreateCommentHandler).Methods("POST")
	auth.HandleFunc("/posts/{postID}/like", postHandlers.LikePostHandler).Methods("POST")
	auth.HandleFunc("/comments/{commentID}/like", postHandlers.LikeCommentHandler).Methods("POST")

	// --- Chat Routes ---
	auth.HandleFunc("/chats/private/{userID}", chatHandlers.GetPrivateConversationHandler).Methods("GET")
	auth.HandleFunc("/chats/group/{groupID}", chatHandlers.GetGroupConversationHandler).Methods("GET")

	// Wrap the router with CORS middleware before returning
	return CORSMiddleware(router)
}