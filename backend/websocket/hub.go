package websocket

import (
	"encoding/json"
	"log"
	"time"

	"social-network/database/models"
)

type RoutedMessage struct {
	Client  *Client
	Message IncomingMessage
}

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

func (h *Hub) handlePrivateMessage(routedMsg *RoutedMessage) {
	senderID := routedMsg.Client.UserID
	recipientID := routedMsg.Message.RecipientID
	content := routedMsg.Message.Content

	canMessage, err := models.CanUsersMessage(senderID, recipientID)
	if err != nil {
		log.Printf("Error checking message permissions for %s -> %s: %v", senderID, recipientID, err)
		return
	}
	if !canMessage {
		log.Printf("Permission denied: User %s cannot message User %s.", senderID, recipientID)
		return
	}

	dbMsg := &models.Message{
		SenderID:    senderID,
		RecipientID: recipientID,
		Content:     content,
	}
	if err := models.SaveMessage(dbMsg); err != nil {
		log.Printf("Failed to save private message to DB: %v", err)
		return
	}
	outgoingMsg := OutgoingMessage{
		Type:    "private_message",
		Payload: *dbMsg,
	}
	messageBytes, _ := json.Marshal(outgoingMsg)

	shouldReceiveInstant, err := models.ShouldReceiveInstantMessage(senderID, recipientID)
	if err != nil {
		log.Printf("Error checking instant message permissions for %s -> %s: %v", senderID, recipientID, err)
		shouldReceiveInstant = false
	}

	if shouldReceiveInstant {
		if recipientClients, ok := h.clients[recipientID]; ok {
			for client := range recipientClients {
				client.send <- messageBytes
			}
			log.Printf("Sent instant private message from %s to %s", senderID, recipientID)
		} else {
			log.Printf("Recipient %s is not online. Message saved to DB (instant delivery allowed).", recipientID)
		}
	} else {
		log.Printf("Recipient %s will not receive instant message from %s (no follow relationship and private profile). Message saved to DB.", recipientID, senderID)
	}

	if senderClients, ok := h.clients[senderID]; ok {
		for client := range senderClients {
			client.send <- messageBytes
		}
	}
}

func (h *Hub) handleGroupMessage(routedMsg *RoutedMessage) {
	senderID := routedMsg.Client.UserID
	groupID := routedMsg.Message.GroupID
	content := routedMsg.Message.Content
	isMember, err := models.IsUserInGroup(senderID, groupID)
	if err != nil {
		log.Printf("Error checking group membership for user %s in group %s: %v", senderID, groupID, err)
		return
	}
	if !isMember {
		log.Printf("Permission denied: User %s is not in group %s.", senderID, groupID)
		return
	}

	dbMsg := &models.Message{
		SenderID: senderID,
		GroupID:  groupID,
		Content:  content,
	}
	if err := models.SaveMessage(dbMsg); err != nil {
		log.Printf("Failed to save group message to DB: %v", err)
		return
	}

	outgoingMsg := OutgoingMessage{
		Type:    "group_message",
		Payload: *dbMsg,
	}
	messageBytes, _ := json.Marshal(outgoingMsg)

	memberIDs, err := models.GetGroupMemberIDs(groupID)
	if err != nil {
		log.Printf("Failed to get group members for group %s: %v", groupID, err)
		return
	}

	for _, memberID := range memberIDs {
		if memberClients, ok := h.clients[memberID]; ok {
			for client := range memberClients {
				client.send <- messageBytes
			}
		}
	}
	log.Printf("Broadcast group message from %s to group %s", senderID, groupID)
}

func (h *Hub) SendNotification(userID, actorID, notifType, message string) {
	notification := &models.Notification{
		UserID:  userID,
		ActorID: actorID,
		Type:    notifType,
		Message: message,
		Read:    false,
	}

	if err := models.CreateNotification(notification); err != nil {
		log.Printf("Failed to save notification to DB: %v", err)
		return
	}

	if userClients, ok := h.clients[userID]; ok {
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

		for client := range userClients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(userClients, client)
			}
		}
		log.Printf("Sent real-time notification of type '%s' to user %s", notifType, userID)
	}
}

func (h *Hub) SendFollowRequestUpdate(userID string) {
	if userClients, ok := h.clients[userID]; ok {
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

		for client := range userClients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(userClients, client)
			}
		}
		log.Printf("Sent follow request update to user %s", userID)
	}
}

func (h *Hub) SendUserListUpdate(userID string) {
	if userClients, ok := h.clients[userID]; ok {
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

		for client := range userClients {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(userClients, client)
			}
		}
		log.Printf("Sent user list update to user %s", userID)
	}
}

func (h *Hub) SendMessageToUser(userID string, message *models.Message) {
	if userClients, ok := h.clients[userID]; ok {
		messagePayload := struct {
			Type string          `json:"type"`
			Data *models.Message `json:"data"`
		}{
			Type: "message",
			Data: message,
		}

		jsonData, err := json.Marshal(messagePayload)
		if err != nil {
			log.Printf("Error marshaling message to JSON: %v", err)
			return
		}

		for client := range userClients {
			select {
			case client.send <- jsonData:
			default:
				close(client.send)
				delete(userClients, client)
				if len(userClients) == 0 {
					delete(h.clients, userID)
				}
			}
		}
	}
}
