package models

import (
	"time"
	"social-network/database"

	"github.com/google/uuid"
)

type Message struct {
	ID          string
	SenderID    string
	RecipientID string // Can be empty for group messages
	GroupID     string // Can be empty for private messages
	Content     string
	CreatedAt   time.Time
}

// SaveMessage stores a new chat message in the database.
func SaveMessage(msg *Message) error {
	msg.ID = uuid.NewString() // Generate a new ID for the message
	stmt, err := database.DB.Prepare(`
		INSERT INTO chat_messages (id, sender_id, recipient_id, group_id, content)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(msg.ID, msg.SenderID, msg.RecipientID, msg.GroupID, msg.Content)
	return err
}