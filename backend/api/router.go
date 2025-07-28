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

	// --- WebSocket Route ---
	apiRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// --- Protected Routes (require a valid session cookie via AuthMiddleware) ---
	auth := apiRouter.PathPrefix("").Subrouter()


	auth.Use(AuthMiddleware)



	// --- User & Follower Routes ---
	auth.HandleFunc("/me", userHandlers.CurrentUserHandler).Methods("GET")
	auth.HandleFunc("/follow/{userId}", userHandlers.SendFollowRequestHandler).Methods("POST")
	auth.HandleFunc("/follow/{userId}/accept", userHandlers.AcceptFollowRequestHandler).Methods("POST")
	auth.HandleFunc("/follow/{userId}/decline", userHandlers.DeclineFollowRequestHandler).Methods("POST")
	auth.HandleFunc("/unfollow/{userId}", userHandlers.UnfollowHandler).Methods("POST")
	auth.HandleFunc("/followers/{userId}", userHandlers.ListFollowersHandler).Methods("GET")
	auth.HandleFunc("/following/{userId}", userHandlers.ListFollowingHandler).Methods("GET")
	auth.HandleFunc("/follow-requests", userHandlers.ListPendingFollowRequestsHandler).Methods("GET")

	// --- Profile Routes ---
	auth.HandleFunc("/profile/{userId}", userHandlers.GetProfileHandler).Methods("GET")
	auth.HandleFunc("/profile", userHandlers.UpdateProfileHandler).Methods("PUT")
	auth.HandleFunc("/profile/avatar", userHandlers.UploadAvatarHandler).Methods("POST")
	auth.HandleFunc("/profile/make-private", userHandlers.MakeProfilePrivateHandler).Methods("POST")

	// --- Post & Feed Routes ---
	auth.HandleFunc("/posts", postHandlers.CreatePostHandler).Methods("POST")
	auth.HandleFunc("/posts/feed", postHandlers.GetFeedPostsHandler).Methods("GET")
	auth.HandleFunc("/posts/{postID}/comment", postHandlers.CreateCommentHandler).Methods("POST")

	// --- NEW LIKE/DISLIKE ROUTES ---
	auth.HandleFunc("/posts/{postID}/like", postHandlers.LikePostHandler).Methods("POST")
	auth.HandleFunc("/comments/{commentID}/like", postHandlers.LikeCommentHandler).Methods("POST")

	// Wrap the router with CORS middleware before returning
	return CORSMiddleware(router)
}