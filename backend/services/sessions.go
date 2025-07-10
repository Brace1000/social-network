package services

import (
	"net/http"
	"time"

	"social-network/database/models"

	"github.com/google/uuid"
)

type contextKey string

const (
	SessionCookieName = "social_network_session"
	SessionDuration   = 24 * 7 * time.Hour // Sessions last for one week
	UserContextKey    = contextKey("user")
)

// CreateSession creates a new session for a user and returns the session token.
func CreateSession(userID string) (string, error) {
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(SessionDuration)

	session := &models.Session{
		Token:     sessionToken,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	err := models.CreateSession(session)
	if err != nil {
		return "", err
	}
	return sessionToken, nil
}

// SetSessionCookie sets the session cookie on the HTTP response.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Expires:  time.Now().Add(SessionDuration),
		HttpOnly: true,
		Path:     "/",
	})
}

// GetUserFromSession retrieves the user associated with a session token from a request cookie.
func GetUserFromSession(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, nil // No cookie means no user is logged in
	}

	sessionToken := cookie.Value
	session, err := models.GetSessionByToken(sessionToken)
	if err != nil {
		return nil, err // Database error
	}

	if session == nil || session.IsExpired() {
		return nil, nil // Session not found or expired
	}

	user, err := models.GetUserByID(session.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ClearSessionCookie logs the user out by deleting the session and expiring the cookie.
func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return
	}
	models.DeleteSession(cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})
}
