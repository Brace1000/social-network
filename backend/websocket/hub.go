package websocket

import (
	"encoding/json"
	"log"
	"time"

	"social-network/database/models"
)

// RoutedMessage wraps an incoming message with the client who sent it.
type RoutedMessage struct {
	Client  *Client
	Message IncomingMessage
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients      map[string]map[*Client]bool
	routeMessage chan *RoutedMessage
	register     chan *Client
	unregister   chan *Client
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
			case "group_message":
				h.handleGroupMessage(routedMsg)
			default:
				log.Printf("Unknown message type: %s", routedMsg.Message.Type)
			}
		}
	}
}

// handlePrivateMessage processes and routes a 1-to-1 message.
func (h *Hub) handlePrivateMessage(routedMsg *RoutedMessage) {
	senderID := routedMsg.Client.UserID
	recipientID := routedMsg.Message.RecipientID
	content := routedMsg.Message.Content

	// AUDIT POINT: Check if users are allowed to message each other.
	canMessage, err := models.CanUsersMessage(senderID, recipientID)
	if err != nil {
		log.Printf("Error checking message permissions for %s -> %s: %v", senderID, recipientID, err)
		return
	}
	if !canMessage {
		log.Printf("Permission denied: User %s cannot message User %s.", senderID, recipientID)
		// Optionally, send an error message back to the sender.
		return
	}

	// Persist the message to the database.
	dbMsg := &models.Message{
		SenderID:    senderID,
		RecipientID: recipientID,
		Content:     content,
	}
	if err := models.SaveMessage(dbMsg); err != nil {
		log.Printf("Failed to save private message to DB: %v", err)
		return
	}

	// Create the outgoing message payload.
	outgoingMsg := OutgoingMessage{
		Type:    "private_message",
		Payload: *dbMsg, // Send the full message object with ID and timestamp
	}
	messageBytes, _ := json.Marshal(outgoingMsg)

	// AUDIT POINT: Send the message ONLY to the recipient's clients.
	if recipientClients, ok := h.clients[recipientID]; ok {
		for client := range recipientClients {
			client.send <- messageBytes
		}
		log.Printf("Sent private message from %s to %s", senderID, recipientID)
	} else {
		log.Printf("Recipient %s is not online. Message saved to DB.", recipientID)
	}

	// Also send the message back to the sender's other devices.
	if senderClients, ok := h.clients[senderID]; ok {
		for client := range senderClients {
			// Don't resend to the originating client tab, though it's often harmless.
			// if client != routedMsg.Client {
			client.send <- messageBytes
			// }
		}
	}
}

// handleGroupMessage processes and routes a group message.
func (h *Hub) handleGroupMessage(routedMsg *RoutedMessage) {
	senderID := routedMsg.Client.UserID
	groupID := routedMsg.Message.GroupID
	content := routedMsg.Message.Content

	// AUDIT POINT: Check if the sender is a member of the group.
	isMember, err := models.IsUserInGroup(senderID, groupID)
	if err != nil {
		log.Printf("Error checking group membership for user %s in group %s: %v", senderID, groupID, err)
		return
	}
	if !isMember {
		log.Printf("Permission denied: User %s is not in group %s.", senderID, groupID)
		return
	}

	// Persist the message.
	dbMsg := &models.Message{
		SenderID: senderID,
		GroupID:  groupID,
		Content:  content,
	}
	if err := models.SaveMessage(dbMsg); err != nil {
		log.Printf("Failed to save group message to DB: %v", err)
		return
	}

	// Create the outgoing payload.
	outgoingMsg := OutgoingMessage{
		Type:    "group_message",
		Payload: *dbMsg,
	}
	messageBytes, _ := json.Marshal(outgoingMsg)

	// Get all members of the group to broadcast the message.
	memberIDs, err := models.GetGroupMemberIDs(groupID)
	if err != nil {
		log.Printf("Failed to get group members for group %s: %v", groupID, err)
		return
	}

	// AUDIT POINT: Broadcast to all online members of the group.
	for _, memberID := range memberIDs {
		if memberClients, ok := h.clients[memberID]; ok {
			for client := range memberClients {
				client.send <- messageBytes
			}
		}
	}
	log.Printf("Broadcast group message from %s to group %s", senderID, groupID)
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