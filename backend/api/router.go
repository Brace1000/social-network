package api

import (
	  "net/http"
	"social-network/websocket"

	"github.com/gorilla/mux"
)

// SetupRouter configures all the API routes for the application.
func SetupRouter(hub *websocket.Hub) *mux.Router {
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Authentication routes (public)
	apiRouter.HandleFunc("/register", RegisterHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/logout", LogoutHandler).Methods("POST", "OPTIONS")

	// Route to check current user session (protected)
	apiRouter.HandleFunc("/me", AuthMiddleware(CurrentUserHandler)).Methods("GET", "OPTIONS")

	// WebSocket connection route
	apiRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	return router
}
