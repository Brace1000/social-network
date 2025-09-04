package api

import (
	"net/http"

	"social-network/websocket"

	"github.com/gorilla/mux"
)

func SetupRouter(hub *websocket.Hub) http.Handler {
	userHandlers := NewUserHandlers(hub)
	postHandlers := NewPostHandlers()
	chatHandlers := NewChatHandlers(hub)

	router := mux.NewRouter()

	router.Use(CORSMiddleware)
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/register", userHandlers.RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", userHandlers.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/logout", userHandlers.LogoutHandler).Methods("POST", "OPTIONS")

	apiRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	auth := apiRouter.PathPrefix("").Subrouter()
	auth.Use(AuthMiddleware)

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

	// Notification Routes
	auth.HandleFunc("/notifications", userHandlers.GetNotificationsHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/notifications/{notificationId}/read", userHandlers.MarkNotificationAsReadHandler).Methods("POST", "OPTIONS")

	// Profile Routes
	auth.HandleFunc("/profile/{userId}", userHandlers.GetProfileHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/profile", userHandlers.UpdateProfileHandler).Methods("PUT", "OPTIONS")
	auth.HandleFunc("/profile/avatar", userHandlers.UploadAvatarHandler).Methods("POST", "OPTIONS")
	auth.HandleFunc("/profile/toggle-privacy", userHandlers.ToggleProfilePrivacyHandler).Methods("POST", "OPTIONS")

	// Post & Feed Routes
	auth.HandleFunc("/posts", postHandlers.CreatePostHandler).Methods("POST")
	auth.HandleFunc("/posts/feed", postHandlers.GetFeedPostsHandler).Methods("GET")
	auth.HandleFunc("/posts/{postID}/comment", postHandlers.CreateCommentHandler).Methods("POST")
	auth.HandleFunc("/posts/{postID}/like", postHandlers.LikePostHandler).Methods("POST")
	auth.HandleFunc("/comments/{commentID}/like", postHandlers.LikeCommentHandler).Methods("POST")

	// Chat Routes
	auth.HandleFunc("/chats/conversations", chatHandlers.GetConversationsHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/chats/private/{userID}", chatHandlers.GetPrivateConversationHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/chats/group/{groupID}", chatHandlers.GetGroupConversationHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/chats/can-message/{userID}", chatHandlers.CheckCanMessageHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/chats/search-users", chatHandlers.SearchUsersHandler).Methods("GET", "OPTIONS")
	auth.HandleFunc("/chats/send", chatHandlers.SendMessageHandler).Methods("POST", "OPTIONS")

	return router
}
