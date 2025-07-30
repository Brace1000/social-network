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
		// In production, you should validate the origin
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

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	user, err := services.GetUserFromSession(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Unauthorized WebSocket connection attempt")
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create client
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		UserID: user.ID,
	}
	client.hub.register <- client

	// Start communication
	go client.writePump()
	go client.readPump()
}