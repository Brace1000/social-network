package services

import (
	"io"
	"net/http"
	"os"
	"time"

	"social-network/database/models"

	"github.com/google/uuid"
)

type contextKey string

const (
	SessionCookieName = "social_network_session"
	SessionDuration   = 24 * 7 * time.Hour
	UserContextKey    = contextKey("user")
)

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

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Expires:  time.Now().Add(SessionDuration),
		HttpOnly: true,
		Path:     "/",
	})
}

func GetUserFromSession(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, nil
	}

	sessionToken := cookie.Value
	return GetUserFromSessionToken(sessionToken)
}

func GetUserFromSessionToken(sessionToken string) (*models.User, error) {
	session, err := models.GetSessionByToken(sessionToken)
	if err != nil {
		return nil, err
	}

	if session == nil || session.IsExpired() {
		return nil, nil
	}

	user, err := models.GetUserByID(session.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

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

func SaveUploadedFile(src io.Reader, dstPath string) (*os.File, error) {
	dir := "./uploads/avatars"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(f, src)
	if err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
