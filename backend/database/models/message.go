package models

import (
	"database/sql"
	"social-network/database"
	"time"

	"github.com/google/uuid"
)

// Message represents a single chat message, for both private and group chats.
type Message struct {
	ID          string    `json:"id"`
	SenderID    string    `json:"senderId"`
	RecipientID string    `json:"recipientId,omitempty"` // Empty for group messages
	GroupID     string    `json:"groupId,omitempty"`     // Empty for private messages
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
}

// SaveMessage stores a new chat message in the database.
func SaveMessage(msg *Message) error {
	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now() // Ensure timestamp is set before saving

	// Use sql.NullString for optional fields to handle NULL values correctly.
	var recipient sql.NullString
	if msg.RecipientID != "" {
		recipient.String = msg.RecipientID
		recipient.Valid = true
	}
	var group sql.NullString
	if msg.GroupID != "" {
		group.String = msg.GroupID
		group.Valid = true
	}

	stmt, err := database.DB.Prepare(`
		INSERT INTO chat_messages (id, sender_id, recipient_id, group_id, content, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(msg.ID, msg.SenderID, recipient, group, msg.Content, msg.CreatedAt)
	return err
}

// CanUsersMessage checks if two users are allowed to chat.
// This directly addresses the audit requirement.
func CanUsersMessage(senderID, recipientID string) (bool, error) {
	// A user can always message themselves.
	if senderID == recipientID {
		return true, nil
	}

	// Logic: A chat is allowed if they follow each other.
	var count int
	query := `
		SELECT COUNT(*) FROM followers f1
		JOIN followers f2 ON f1.follower_id = f2.followed_id AND f1.followed_id = f2.follower_id
		WHERE f1.follower_id = ? AND f1.followed_id = ? AND f1.status = 'accepted' AND f2.status = 'accepted'
	`
	err := database.DB.QueryRow(query, senderID, recipientID).Scan(&count)
	if err != nil {
		return false, err
	}

	// NOTE: You could add logic here to check if the recipient's profile is public.
	

	return count > 0, nil
}

// IsUserInGroup checks if a user is a member of a group.

func IsUserInGroup(userID, groupID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM group_members WHERE user_id = ? AND group_id = ?)"
	err := database.DB.QueryRow(query, userID, groupID).Scan(&exists)
	return exists, err
}

// GetGroupMemberIDs returns a slice of all user IDs in a given group.
func GetGroupMemberIDs(groupID string) ([]string, error) {
	rows, err := database.DB.Query("SELECT user_id FROM group_members WHERE group_id = ?", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberIDs []string
	for rows.Next() {
		var memberID string
		if err := rows.Scan(&memberID); err != nil {
			return nil, err
		}
		memberIDs = append(memberIDs, memberID)
	}
	return memberIDs, nil
}