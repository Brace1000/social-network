package models

import (
	"database/sql"
	"social-network/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupNotifTestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	database.DB = db
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			first_name TEXT
		);
		CREATE TABLE notifications (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			actor_id TEXT,
			type TEXT,
			message TEXT,
			read INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (id, first_name) VALUES ('u1', 'Alice'), ('u2', 'Bob')`)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}
}

func TestCreateAndGetNotifications(t *testing.T) {
	setupNotifTestDB(t)
	notif := &Notification{
		UserID:  "u1",
		ActorID: "u2",
		Type:    "follow_request",
		Message: "Bob wants to follow you.",
		Read:    false,
	}
	err := CreateNotification(notif)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}
	notifs, err := GetNotificationsForUser("u1")
	if err != nil {
		t.Fatalf("GetNotificationsForUser failed: %v", err)
	}
	if len(notifs) != 1 || notifs[0].Message != "Bob wants to follow you." {
		t.Fatalf("GetNotificationsForUser did not return correct notification: %+v", notifs)
	}
	// No notifications for u2
	notifs, err = GetNotificationsForUser("u2")
	if err != nil {
		t.Fatalf("GetNotificationsForUser (empty) failed: %v", err)
	}
	if len(notifs) != 0 {
		t.Fatalf("GetNotificationsForUser (empty) should be 0, got: %d", len(notifs))
	}
}
