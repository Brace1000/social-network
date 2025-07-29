package websocket

import (
	"log"
	"net/http"
	"social-network/services"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		
		return true 
	},
}
// getSessionTokenFromRequest extracts the session token from either cookie or query parameter
func getSessionTokenFromRequest(r *http.Request) string {
	// First try to get from cookie
	if cookie, err := r.Cookie(services.SessionCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	
	// If not in cookie, try query parameter
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	
	return ""
}

// ServeWs authenticates the user and handles the websocket connection.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate the user from the session cookie
	user, err := services.GetUserFromSession(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Unauthorized WebSocket connection attempt.")
		return
	}

	// 2. Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 3. Create a new client with the authenticated UserID
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		UserID: user.ID,
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the goroutines
	go client.writePump()
	go client.readPump()
}