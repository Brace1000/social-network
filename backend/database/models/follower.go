package models

import (
	"social-network/database"
	"database/sql"
	"github.com/google/uuid"
)

// AreFollowing checks if a one-way follow relationship exists.
// Specifically, it checks if user1 is following user2.
func AreFollowing(user1ID, user2ID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)"
	err := database.DB.QueryRow(query, user1ID, user2ID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CheckFollowRelationship verifies if two users can message each other.
// This is true if user1 follows user2, OR user2 follows user1.
func CheckFollowRelationship(user1ID, user2ID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM followers WHERE (follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)
		)
	`
	err := database.DB.QueryRow(query, user1ID, user2ID, user2ID, user1ID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CreateFollowRequest creates a new follow request (pending) from requester to target.
func CreateFollowRequest(requesterID, targetID string) error {
	id := uuid.NewString()
	_, err := database.DB.Exec(
		`INSERT INTO follow_requests (id, requester_id, target_id, status) VALUES (?, ?, ?, 'pending')`,
		id, requesterID, targetID,
	)
	return err
}

// AcceptFollowRequest accepts a pending follow request and adds to followers.
func AcceptFollowRequest(requesterID, targetID string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`UPDATE follow_requests SET status = 'accepted' WHERE requester_id = ? AND target_id = ? AND status = 'pending'`, requesterID, targetID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT OR IGNORE INTO followers (follower_id, following_id) VALUES (?, ?)`, requesterID, targetID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// DeclineFollowRequest declines a pending follow request.
func DeclineFollowRequest(requesterID, targetID string) error {
	_, err := database.DB.Exec(`UPDATE follow_requests SET status = 'declined' WHERE requester_id = ? AND target_id = ? AND status = 'pending'`, requesterID, targetID)
	return err
}

// RemoveFollower removes a follower relationship.
func RemoveFollower(followerID, followingID string) error {
	_, err := database.DB.Exec(`DELETE FROM followers WHERE follower_id = ? AND following_id = ?`, followerID, followingID)
	return err
}

// ListFollowers returns a list of user IDs who follow the given user.
func ListFollowers(userID string) ([]string, error) {
	rows, err := database.DB.Query(`SELECT follower_id FROM followers WHERE following_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var followers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		followers = append(followers, id)
	}
	return followers, nil
}

// ListFollowing returns a list of user IDs whom the given user is following.
func ListFollowing(userID string) ([]string, error) {
	rows, err := database.DB.Query(`SELECT following_id FROM followers WHERE follower_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var following []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		following = append(following, id)
	}
	return following, nil
}

// ListPendingFollowRequests returns pending follow requests for a user (as target).
func ListPendingFollowRequests(targetID string) ([]string, error) {
	rows, err := database.DB.Query(`SELECT requester_id FROM follow_requests WHERE target_id = ? AND status = 'pending'`, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var requesters []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		requesters = append(requesters, id)
	}
	return requesters, nil
}
