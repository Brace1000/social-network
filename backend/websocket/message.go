package websocket

import "time"

// IncomingMessage is the format for messages received from clients.
type IncomingMessage struct {
	Type        string `json:"type"`        // "private_message", "group_message"
	RecipientID string `json:"recipientId,omitempty"` // UserID for private, GroupID for group
	Content     string `json:"content"`
}

// OutgoingMessage is the format for messages sent back to clients.
type OutgoingMessage struct {
	Type      string    `json:"type"`
	SenderID  string    `json:"senderId"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
type NotificationMessage struct {
	Type      string    `json:"type"` // "notification"
	Payload   struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		ActorID   string `json:"actorId,omitempty"`
		NotifType string `json:"notifType"` // "follow_request", etc.
		Read      bool   `json:"read"`
	} `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}