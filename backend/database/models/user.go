package models

import (
	"database/sql"
	"social-network/database"

	"time"
)

// User represents the structure of the 'users' table.
type User struct {
	ID           string
	FirstName    string
	LastName     string
	Nickname     string
	Email        string
	PasswordHash string
	DateOfBirth  string
	AvatarPath   string
	AboutMe      string
	IsPublic     bool
	CreatedAt    time.Time
}

// CreateUser inserts a new user into the database.
func CreateUser(user *User) error {
    stmt, err := database.DB.Prepare(`
        INSERT INTO users (id, first_name, last_name, nickname, email, password_hash, date_of_birth, avatar_path, about_me, is_public)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    // Handle optional fields that can be NULL in the database
    var nickname, avatar, aboutMe interface{}
    if user.Nickname != "" {
        nickname = user.Nickname
    }
    if user.AvatarPath != "" {
        avatar = user.AvatarPath
    }
    if user.AboutMe != "" {
        aboutMe = user.AboutMe
    }

    _, err = stmt.Exec(
        user.ID,
        user.FirstName,
        user.LastName,
        nickname, 
        user.Email,
        user.PasswordHash,
        user.DateOfBirth,
        avatar,   
        aboutMe,  
        user.IsPublic,
    )
    return err
}

// GetUserByEmail retrieves a user by their email address. Returns nil if no user is found.
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	row := database.DB.QueryRow("SELECT id, first_name, last_name, nickname, email, password_hash, date_of_birth, avatar_path, about_me, is_public, created_at FROM users WHERE email = ?", email)

	var avatar sql.NullString
	var nickname sql.NullString
	var aboutMe sql.NullString

	err := row.Scan(
		&user.ID, &user.FirstName, &user.LastName, &nickname, &user.Email, &user.PasswordHash,
		&user.DateOfBirth, &avatar, &aboutMe, &user.IsPublic, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No user found is not an application error
		}
		return nil, err // A real database error occurred
	}

	user.AvatarPath = avatar.String
	user.Nickname = nickname.String
	user.AboutMe = aboutMe.String

	return user, nil
}

// GetUserByID retrieves a user by their unique ID. Returns nil if no user is found.
func GetUserByID(id string) (*User, error) {
	user := &User{}
	row := database.DB.QueryRow("SELECT id, first_name, last_name, nickname, email, password_hash, date_of_birth, avatar_path, about_me, is_public, created_at FROM users WHERE id = ?", id)

	var avatar sql.NullString
	var nickname sql.NullString
	var aboutMe sql.NullString

	err := row.Scan(
		&user.ID, &user.FirstName, &user.LastName, &nickname, &user.Email, &user.PasswordHash,
		&user.DateOfBirth, &avatar, &aboutMe, &user.IsPublic, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	user.AvatarPath = avatar.String
	user.Nickname = nickname.String
	user.AboutMe = aboutMe.String

	return user, nil
}
