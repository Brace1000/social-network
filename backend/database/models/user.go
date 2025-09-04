package models

import (
	"database/sql"
	"time"

	"social-network/database"
)

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

func CreateUser(user *User) error {
	stmt, err := database.DB.Prepare(`
                                       INSERT INTO users (id, first_name, last_name, nickname, email, password_hash, date_of_birth, 
	                                                  	avatar_path, about_me, is_public)
                                        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

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

func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	row := database.DB.QueryRow(`
	                             SELECT id, first_name, last_name, nickname, email, 
	                                        password_hash, date_of_birth, avatar_path, about_me, is_public
	                             FROM users WHERE email = ?`, email)

	var avatar sql.NullString
	var nickname sql.NullString
	var aboutMe sql.NullString

	err := row.Scan(
		&user.ID, &user.FirstName, &user.LastName, &nickname, &user.Email, &user.PasswordHash,
		&user.DateOfBirth, &avatar, &aboutMe, &user.IsPublic,
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

func GetUserByID(id string) (*User, error) {
	user := &User{}
	row := database.DB.QueryRow(`
	                           SELECT id, first_name, last_name, nickname, email, password_hash, date_of_birth, avatar_path, about_me, is_public
	                            FROM users WHERE id = ?`, id)

	var avatar sql.NullString
	var nickname sql.NullString
	var aboutMe sql.NullString

	err := row.Scan(
		&user.ID, &user.FirstName, &user.LastName, &nickname, &user.Email, &user.PasswordHash,
		&user.DateOfBirth, &avatar, &aboutMe, &user.IsPublic,
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

func SetUserProfilePrivacy(userID string, isPublic bool) error {
	stmt, err := database.DB.Prepare("UPDATE users SET is_public = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(isPublic, userID)
	return err
}

func UpdateUserProfile(user *User) error {
	stmt, err := database.DB.Prepare(
		`UPDATE users SET first_name = ?, last_name = ?, nickname = ?, about_me = ?, 
		                           is_public = ?, avatar_path = ? WHERE id = ?`)
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

func GetAllUsers() ([]*User, error) {
	rows, err := database.DB.Query(
		`SELECT id, first_name, last_name, nickname, email,
	                                            password_hash, date_of_birth, avatar_path, about_me, is_public, created_at 
                                 	FROM users`)
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
