package models

import (
	"database/sql"
	"time"

	"social-network/database"
)

type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

func CreateSession(session *Session) error {
	stmt, err := database.DB.Prepare("INSERT INTO sessions (token, user_id, expiry) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(session.Token, session.UserID, session.ExpiresAt)
	return err
}

func GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	err := database.DB.QueryRow("SELECT token, user_id, expiry FROM sessions WHERE token = ?", token).
		Scan(&session.Token, &session.UserID, &session.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}

func DeleteSession(token string) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}
