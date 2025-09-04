package models

import (
	"database/sql"
	// "os"
	"social-network/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) {
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
		CREATE TABLE followers (
			follower_id TEXT,
			following_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, following_id)
		);
		CREATE TABLE follow_requests (
			id TEXT PRIMARY KEY,
			requester_id TEXT,
			target_id TEXT,
			status TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}
}

func TestFollowRequestLifecycle(t *testing.T) {
	setupTestDB(t)

	_, err := database.DB.Exec(
		`INSERT INTO users (id, first_name, last_name, is_public) VALUES ('u1', 'Alice', 'A', 0), ('u2', 'Bob', 'B', 1)`)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}

	err = CreateFollowRequest("u1", "u2")
	if err != nil {
		t.Fatalf("CreateFollowRequest failed: %v", err)
	}

	reqs, err := ListPendingFollowRequests("u2")
	if err != nil || len(reqs) != 1 || reqs[0] != "u1" {
		t.Fatalf("ListPendingFollowRequests failed: %v, got: %v", err, reqs)
	}

	err = AcceptFollowRequest("u1", "u2")
	if err != nil {
		t.Fatalf("AcceptFollowRequest failed: %v", err)
	}

	isFollowing, err := AreFollowing("u1", "u2")
	if err != nil || !isFollowing {
		t.Fatalf("AreFollowing failed: %v, got: %v", err, isFollowing)
	}

	followers, err := ListFollowers("u2")
	if err != nil || len(followers) != 1 || followers[0] != "u1" {
		t.Fatalf("ListFollowers failed: %v, got: %v", err, followers)
	}

	following, err := ListFollowing("u1")
	if err != nil || len(following) != 1 || following[0] != "u2" {
		t.Fatalf("ListFollowing failed: %v, got: %v", err, following)
	}

	err = RemoveFollower("u1", "u2")
	if err != nil {
		t.Fatalf("RemoveFollower failed: %v", err)
	}
	isFollowing, err = AreFollowing("u1", "u2")
	if err != nil || isFollowing {
		t.Fatalf("AreFollowing after remove failed: %v, got: %v", err, isFollowing)
	}

	err = CreateFollowRequest("u1", "u2")
	if err != nil {
		t.Fatalf("CreateFollowRequest (again) failed: %v", err)
	}
	err = DeclineFollowRequest("u1", "u2")
	if err != nil {
		t.Fatalf("DeclineFollowRequest failed: %v", err)
	}
	reqs, err = ListPendingFollowRequests("u2")
	if err != nil || len(reqs) != 0 {
		t.Fatalf("ListPendingFollowRequests after decline failed: %v, got: %v", err, reqs)
	}
}

func TestCheckFollowRelationship(t *testing.T) {
	setupTestDB(t)
	_, err := database.DB.Exec(`INSERT INTO users (id, first_name, last_name, is_public) VALUES ('a', 'A', 'A', 1), ('b', 'B', 'B', 1)`)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}

	rel, err := CheckFollowRelationship("a", "b")
	if err != nil {
		t.Fatalf("CheckFollowRelationship failed: %v", err)
	}
	if rel {
		t.Fatalf("CheckFollowRelationship should be false")
	}

	_ = CreateFollowRequest("a", "b")
	_ = AcceptFollowRequest("a", "b")
	rel, err = CheckFollowRelationship("a", "b")
	if err != nil || !rel {
		t.Fatalf("CheckFollowRelationship should be true after follow")
	}

	_ = CreateFollowRequest("b", "a")
	_ = AcceptFollowRequest("b", "a")
	rel, err = CheckFollowRelationship("a", "b")
	if err != nil || !rel {
		t.Fatalf("CheckFollowRelationship should be true for mutual follow")
	}
}

func TestDoubleFollowRequest(t *testing.T) {
	setupTestDB(t)
	_, err := database.DB.Exec(`INSERT INTO users (id, first_name, last_name, is_public) VALUES ('x', 'X', 'X', 1), ('y', 'Y', 'Y', 1)`)
	if err != nil {
		t.Fatalf("failed to insert users: %v", err)
	}
	err = CreateFollowRequest("x", "y")
	if err != nil {
		t.Fatalf("CreateFollowRequest failed: %v", err)
	}
	err = CreateFollowRequest("x", "y")
	if err != nil {
		t.Fatalf("CreateFollowRequest (duplicate) failed: %v", err)
	}
	reqs, _ := ListPendingFollowRequests("y")
	count := 0
	for _, r := range reqs {
		if r == "x" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("Duplicate follow request found: %v", count)
	}
}

func TestFollowFunctions_DBError(t *testing.T) {

	setupTestDB(t)
	database.DB.Close()
	if err := CreateFollowRequest("a", "b"); err == nil {
		t.Fatalf("CreateFollowRequest should fail on closed DB")
	}
	if _, err := ListFollowers("a"); err == nil {
		t.Fatalf("ListFollowers should fail on closed DB")
	}
	if _, err := ListFollowing("a"); err == nil {
		t.Fatalf("ListFollowing should fail on closed DB")
	}
	if _, err := ListPendingFollowRequests("a"); err == nil {
		t.Fatalf("ListPendingFollowRequests should fail on closed DB")
	}
}
