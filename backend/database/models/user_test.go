package models

import (
	"database/sql"
	"social-network/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupUserTestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	database.DB = db
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			nickname TEXT,
			email TEXT,
			password_hash TEXT,
			date_of_birth TEXT,
			avatar_path TEXT,
			about_me TEXT,
			is_public INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
}

func TestUserProfileLifecycle(t *testing.T) {
	setupUserTestDB(t)
	user := &User{
		ID:           "u1",
		FirstName:    "Alice",
		LastName:     "A",
		Nickname:     "ali",
		Email:        "alice@example.com",
		PasswordHash: "hash",
		DateOfBirth:  "2000-01-01",
		AvatarPath:   "",
		AboutMe:      "Hello!",
		IsPublic:     true,
	}
	err := CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	got, err := GetUserByID("u1")
	if err != nil || got == nil || got.FirstName != "Alice" {
		t.Fatalf("GetUserByID failed: %v, got: %+v", err, got)
	}

	got.FirstName = "Alicia"
	got.Nickname = "ally"
	got.AboutMe = "Updated!"
	got.IsPublic = false
	got.AvatarPath = "avatar.png"
	err = UpdateUserProfile(got)
	if err != nil {
		t.Fatalf("UpdateUserProfile failed: %v", err)
	}

	updated, err := GetUserByID("u1")
	if err != nil || updated == nil || updated.FirstName != "Alicia" || updated.Nickname != "ally" || updated.AboutMe != "Updated!" || updated.IsPublic != false || updated.AvatarPath != "avatar.png" {
		t.Fatalf("UpdateUserProfile did not persist changes: %+v", updated)
	}
}

func TestGetUserByEmail(t *testing.T) {
	setupUserTestDB(t)
	user := &User{
		ID:           "u2",
		FirstName:    "Bob",
		LastName:     "B",
		Nickname:     "bobby",
		Email:        "bob@example.com",
		PasswordHash: "hash2",
		DateOfBirth:  "1999-12-31",
		AvatarPath:   "",
		AboutMe:      "Hi!",
		IsPublic:     true,
	}
	err := CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	got, err := GetUserByEmail("bob@example.com")
	if err != nil || got == nil || got.ID != "u2" {
		t.Fatalf("GetUserByEmail failed: %v, got: %+v", err, got)
	}

	notfound, err := GetUserByEmail("nope@example.com")
	if err != nil || notfound != nil {
		t.Fatalf("GetUserByEmail notfound failed: %v, got: %+v", err, notfound)
	}
}

func TestSetUserProfilePrivacy(t *testing.T) {
	setupUserTestDB(t)
	user := &User{
		ID:           "u3",
		FirstName:    "Carol",
		LastName:     "C",
		Nickname:     "carol",
		Email:        "carol@example.com",
		PasswordHash: "hash3",
		DateOfBirth:  "1998-11-11",
		IsPublic:     true,
	}
	err := CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	err = SetUserProfilePrivacy("u3", false)
	if err != nil {
		t.Fatalf("SetUserProfilePrivacy failed: %v", err)
	}
	got, err := GetUserByID("u3")
	if err != nil || got == nil || got.IsPublic != false {
		t.Fatalf("SetUserProfilePrivacy did not persist: %+v", got)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	setupUserTestDB(t)
	got, err := GetUserByID("doesnotexist")
	if err != nil {
		t.Fatalf("GetUserByID notfound error: %v", err)
	}
	if got != nil {
		t.Fatalf("GetUserByID notfound should be nil, got: %+v", got)
	}
}
