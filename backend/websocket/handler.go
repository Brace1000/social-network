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