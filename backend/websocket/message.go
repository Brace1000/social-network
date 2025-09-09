package websocket

import ("time"
"social-network/database/models"
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
	Type        string `json:"type"`                  // "private_message" or "group_message"
	RecipientID string `json:"recipientId,omitempty"` // UserID for private messages
	GroupID     string `json:"groupId,omitempty"`     // GroupID for group messages
	Content     string `json:"content"`
}

// OutgoingMessage is the format for messages sent back to clients.
type OutgoingMessage struct {
	Type    string         `json:"type"`    // "private_message" or "group_message"
	Payload models.Message `json:"payload"` // The full message object from the database

}

type NotificationMessage struct {
	Type      string    `json:"type"`
	Payload   struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		ActorID   string `json:"actorId,omitempty"`
		NotifType string `json:"notifType"`
		Read      bool   `json:"read"`
	} `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}