package main

import (
	"log"
	"net/http"

	"social-network/api"
	"social-network/database"
	"social-network/websocket"
)

func main() {
	// Initialize the database connection and apply migrations
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize WebSocket hub and run it in a separate goroutine
	hub := websocket.NewHub()
	go hub.Run()

	// Setup the API router, passing the hub to it
	router := api.SetupRouter(hub)

	// Start the HTTP server
	log.Println("Server starting on port :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
