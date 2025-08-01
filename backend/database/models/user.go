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
// SetUserProfilePrivacy updates the is_public flag for a given user.
func SetUserProfilePrivacy(userID string, isPublic bool) error {
    stmt, err := database.DB.Prepare("UPDATE users SET is_public = ? WHERE id = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(isPublic, userID)
    return err
}

// UpdateUserProfile updates editable fields of a user in the database.
func UpdateUserProfile(user *User) error {
	stmt, err := database.DB.Prepare(`
		UPDATE users SET first_name = ?, last_name = ?, nickname = ?, about_me = ?, is_public = ?, avatar_path = ? WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.AboutMe,
		user.IsPublic,
		user.AvatarPath,
		user.ID,
	)
	return err
}


// GetAllUsers retrieves all users from the database.
func GetAllUsers() ([]*User, error) {
	rows, err := database.DB.Query("SELECT id, first_name, last_name, nickname, email, password_hash, date_of_birth, avatar_path, about_me, is_public, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var avatar sql.NullString
		var nickname sql.NullString
		var aboutMe sql.NullString
		err := rows.Scan(
			&user.ID, &user.FirstName, &user.LastName, &nickname, &user.Email, &user.PasswordHash,
			&user.DateOfBirth, &avatar, &aboutMe, &user.IsPublic, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		user.AvatarPath = avatar.String
		user.Nickname = nickname.String
		user.AboutMe = aboutMe.String
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}