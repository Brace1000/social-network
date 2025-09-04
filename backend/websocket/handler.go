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

func getSessionTokenFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie(services.SessionCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	return ""
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	sessionToken := getSessionTokenFromRequest(r)
	if sessionToken == "" {
		http.Error(w, "Unauthorized: No session token", http.StatusUnauthorized)
		log.Println("Unauthorized WebSocket connection attempt: No session token")
		return
	}

	user, err := services.GetUserFromSessionToken(sessionToken)
	if err != nil {
		http.Error(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
		log.Printf("Unauthorized WebSocket connection attempt: Error getting user from session token %s: %v", sessionToken, err)
		return
	}
	if user == nil {
		http.Error(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
		log.Printf("Unauthorized WebSocket connection attempt: No user found for session token %s", sessionToken)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn, // Create client
		send:   make(chan []byte, 256),
		UserID: user.ID,
	}
	client.hub.register <- client

	go client.writePump() // Start communication
	go client.readPump()
}
