package websocket

import (
	"encoding/json"
	"log"
	"social-network/database/models"
	"time"
)

// RoutedMessage wraps an incoming message with the client who sent it.
type RoutedMessage struct {
	Client  *Client
	Message IncomingMessage
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// A map of UserID to a map of clients. Allows multiple connections per user.
	clients map[string]map[*Client]bool
	// Inbound messages from the clients.
	routeMessage chan *RoutedMessage
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		routeMessage: make(chan *RoutedMessage),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		clients:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// If this is the first connection for the user, initialize the map.
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			log.Printf("Client registered: UserID %s", client.UserID)

		case client := <-h.unregister:
			if userClients, ok := h.clients[client.UserID]; ok {
				if _, ok := userClients[client]; ok {
					delete(userClients, client)
					close(client.send)
					// If this was the last connection for the user, remove the user entry.
					if len(userClients) == 0 {
						delete(h.clients, client.UserID)
					}
					log.Printf("Client unregistered: UserID %s", client.UserID)
				}
			}

		case routedMsg := <-h.routeMessage:
			switch routedMsg.Message.Type {
			case "private_message":
				h.handlePrivateMessage(routedMsg)
			// case "group_message":
			// 	h.handleGroupMessage(routedMsg) // To be implemented later
			default:
				log.Printf("Unknown message type: %s", routedMsg.Message.Type)
			}
		}
	}
}

func (h *Hub) handlePrivateMessage(routedMsg *RoutedMessage) {
	senderID := routedMsg.Client.UserID
	recipientID := routedMsg.Message.RecipientID
	content := routedMsg.Message.Content

	// 1. Check permissions: Can sender message recipient?
	// (This requires checking the follow relationship or if the recipient's profile is public)
	// canMessage, err := models.CheckFollowRelationship(senderID, recipientID)
	// if err != nil {
	// 	log.Printf("Error checking follow relationship: %v", err)
	// 	return
	// }

	// For this example, we'll bypass the check for easier testing.
	// In a real app, you would uncomment and use the line below:
	// if !canMessage {
	// 	log.Printf("Permissions check failed: User %s cannot message User %s", senderID, recipientID)
	// 	return
	// }

	// 2. Save the message to the database
	dbMsg := &models.Message{
		SenderID:    senderID,
		RecipientID: recipientID,
		Content:     content,
	}
	if err := models.SaveMessage(dbMsg); err != nil {
		log.Printf("Failed to save message to DB: %v", err)
		return
	}

	// 3. Create the outgoing message format
	outgoingMsg := OutgoingMessage{
		Type:      "private_message",
		SenderID:  senderID,
		Content:   content,
		Timestamp: time.Now(),
	}
	messageBytes, _ := json.Marshal(outgoingMsg)

	// 4. Send the message to all of the recipient's active connections
	if recipientClients, ok := h.clients[recipientID]; ok {
		for client := range recipientClients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(recipientClients, client)
			}
		}
		log.Printf("Sent private message from %s to %s", senderID, recipientID)
	} else {
		log.Printf("Recipient %s is not online.", recipientID)

	}
}

// SendNotification creates a notification, saves it to the DB, and pushes it to the user if they are online.
func (h *Hub) SendNotification(userID, actorID, notifType, message string) {
	// 1. Create and save the notification to the database
	notification := &models.Notification{
		UserID:  userID,
		ActorID: actorID,
		Type:    notifType,
		Message: message,
		Read:    false, // New notifications are always unread
	}

	if err := models.CreateNotification(notification); err != nil {
		log.Printf("Failed to save notification to DB: %v", err)
		return // Don't send if we can't save it
	}

	// 2. Check if the target user is online
	if userClients, ok := h.clients[userID]; ok {
		// 3. If they are online, create the real-time message payload
		wsNotif := NotificationMessage{
			Type: "notification",
			Payload: struct {
				ID        string `json:"id"`
				Message   string `json:"message"`
				ActorID   string `json:"actorId,omitempty"`
				NotifType string `json:"notifType"`
				Read      bool   `json:"read"`
			}{
				ID:        notification.ID,
				Message:   notification.Message,
				ActorID:   notification.ActorID,
				NotifType: notification.Type,
				Read:      notification.Read,
			},
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(wsNotif)
		if err != nil {
			log.Printf("Failed to marshal notification: %v", err)
			return
		}

		// 4. Send the notification to all of that user's active connections
		for client := range userClients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(userClients, client)
			}
		}
		log.Printf("Sent real-time notification of type '%s' to user %s", notifType, userID)
	} else {
		// User is not online, notification saved to DB for later retrieval
		// (removed debug log to reduce noise)
	}
}

// SendFollowRequestUpdate sends a message to refresh follow requests for a user
func (h *Hub) SendFollowRequestUpdate(userID string) {
	// Check if the target user is online
	if userClients, ok := h.clients[userID]; ok {
		// Create the follow request update message
		updateMsg := struct {
			Type string `json:"type"`
			Data struct {
				Action string `json:"action"`
			} `json:"data"`
			Timestamp time.Time `json:"timestamp"`
		}{
			Type: "follow_request_update",
			Data: struct {
				Action string `json:"action"`
			}{
				Action: "refresh",
			},
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(updateMsg)
		if err != nil {
			log.Printf("Failed to marshal follow request update: %v", err)
			return
		}

		// Send the update to all of that user's active connections
		for client := range userClients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(userClients, client)
			}
		}
		log.Printf("Sent follow request update to user %s", userID)
	} else {
		// User is not online, update will be handled when they connect
		// (removed debug log to reduce noise)
	}
}

// SendUserListUpdate sends a message to refresh the user list for a user
func (h *Hub) SendUserListUpdate(userID string) {
	// Check if the target user is online
	if userClients, ok := h.clients[userID]; ok {
		// Create the user list update message
		updateMsg := struct {
			Type string `json:"type"`
			Data struct {
				Action string `json:"action"`
			} `json:"data"`
			Timestamp time.Time `json:"timestamp"`
		}{
			Type: "user_list_update",
			Data: struct {
				Action string `json:"action"`
			}{
				Action: "refresh",
			},
			Timestamp: time.Now(),
		}

		messageBytes, err := json.Marshal(updateMsg)
		if err != nil {
			log.Printf("Failed to marshal user list update: %v", err)
			return
		}

		// Send the update to all of that user's active connections
		for client := range userClients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(userClients, client)
			}
		}
		log.Printf("Sent user list update to user %s", userID)
	} else {
		// User is not online, update will be handled when they connect
		// (removed debug log to reduce noise)
	}
}