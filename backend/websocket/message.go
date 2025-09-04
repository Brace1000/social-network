package websocket

import (
	"social-network/database/models"
	"time"
)

type MessageType string

const (
	PrivateMessage MessageType = "private_message"
	GroupMessage   MessageType = "group_message"
	Notification   MessageType = "notification"
	ReadReceipt    MessageType = "read_receipt"
	Typing         MessageType = "typing"
)

type IncomingMessage struct {
	Type        string `json:"type"`
	RecipientID string `json:"recipientId,omitempty"`
	GroupID     string `json:"groupId,omitempty"`
	Content     string `json:"content"`
}

type OutgoingMessage struct {
	Type    string         `json:"type"`
	Payload models.Message `json:"payload"`
}

type NotificationMessage struct {
	Type    string `json:"type"`
	Payload struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		ActorID   string `json:"actorId,omitempty"`
		NotifType string `json:"notifType"`
		Read      bool   `json:"read"`
	} `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}
