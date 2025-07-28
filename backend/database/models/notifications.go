package models

import (
	"database/sql"
	"social-network/database" 
	"time"

	"github.com/google/uuid"
)

// Notification represents the structure of the 'notifications' table.
type Notification struct {
	ID        string
	UserID    string
	ActorID   string
	Type      string
	Message   string
	Read      bool
	CreatedAt time.Time
}

// CreateNotification creates and saves a new notification to the database.
func CreateNotification(notif *Notification) error {
	notif.ID = uuid.NewString() // Generate a unique ID

	var actorID sql.NullString
	if notif.ActorID != "" {
		actorID.String = notif.ActorID
		actorID.Valid = true
	}

	stmt, err := database.DB.Prepare(`
		INSERT INTO notifications (id, user_id, actor_id, type, message, read)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(notif.ID, notif.UserID, actorID, notif.Type, notif.Message, notif.Read)
	return err
}

// GetNotificationsForUser fetches all notifications for a specific user.
func GetNotificationsForUser(userID string) ([]Notification, error) {
	query := "SELECT id, user_id, actor_id, type, message, read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC"
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notif Notification
		var actorID sql.NullString

		// This is the clean, correct scan line.
		err := rows.Scan(
			&notif.ID,
			&notif.UserID,
			&actorID,
			&notif.Type,
			&notif.Message,
			&notif.Read,
			&notif.CreatedAt,
		)
		
		if err != nil {
			return nil, err
		}

		if actorID.Valid {
			notif.ActorID = actorID.String
		}

		notifications = append(notifications, notif)
	}
	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read for a specific user.
func MarkNotificationAsRead(notificationID, userID string) error {
	stmt, err := database.DB.Prepare(`
		UPDATE notifications 
		SET read = 1 
		WHERE id = ? AND user_id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(notificationID, userID)
	return err
}

// GetUnreadNotificationCount returns the count of unread notifications for a user.
func GetUnreadNotificationCount(userID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read = 0"
	err := database.DB.QueryRow(query, userID).Scan(&count)
	return count, err
}