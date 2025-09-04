package models

import (
	"database/sql"
	"social-network/database"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID          string    `json:"id"`
	SenderID    string    `json:"senderId"`
	RecipientID string    `json:"recipientId,omitempty"`
	GroupID     string    `json:"groupId,omitempty"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
}

func SaveMessage(msg *Message) error {
	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()

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

func CanUsersMessage(senderID, recipientID string) (bool, error) {
	if senderID == recipientID {
		return true, nil
	}

	var count int
	query := `
		SELECT COUNT(*) FROM followers
		WHERE (follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)
	`
	err := database.DB.QueryRow(query, senderID, recipientID, recipientID, senderID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func ShouldReceiveInstantMessage(senderID, recipientID string) (bool, error) {
	if senderID == recipientID {
		return true, nil
	}
	var recipientFollowsSender int
	followQuery := `
		SELECT COUNT(*) FROM followers
		WHERE follower_id = ? AND following_id = ?
	`
	err := database.DB.QueryRow(followQuery, recipientID, senderID).Scan(&recipientFollowsSender)
	if err != nil {
		return false, err
	}

	if recipientFollowsSender > 0 {
		return true, nil
	}

	var isPublic bool
	publicQuery := `SELECT is_public FROM users WHERE id = ?`
	err = database.DB.QueryRow(publicQuery, recipientID).Scan(&isPublic)
	if err != nil {
		return false, err
	}
	return isPublic, nil
}

func IsUserInGroup(userID, groupID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM group_members WHERE user_id = ? AND group_id = ?)"
	err := database.DB.QueryRow(query, userID, groupID).Scan(&exists)
	return exists, err
}

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
