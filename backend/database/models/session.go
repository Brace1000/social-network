package models

import (
	"database/sql"
	"social-network/database"
	"time"
)

// Session represents the structure of the 'sessions' table.
type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

// CreateSession inserts a new session into the database.
func CreateSession(session *Session) error {
	stmt, err := database.DB.Prepare("INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(session.Token, session.UserID, session.ExpiresAt)
	return err
}

// GetSessionByToken retrieves a session by its token. Returns nil if no session is found.
func GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	err := database.DB.QueryRow("SELECT session_token, user_id, expires_at FROM sessions WHERE session_token = ?", token).
		Scan(&session.Token, &session.UserID, &session.ExpiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

// DeleteSession removes a session from the database (used for logout).
func DeleteSession(token string) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE session_token = ?", token)
	return err
}
