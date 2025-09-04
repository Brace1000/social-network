package models

import (
	"social-network/database"
)

func AreFollowing(user1ID, user2ID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)"
	err := database.DB.QueryRow(query, user1ID, user2ID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

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

func AcceptFollowRequest(requesterID, targetID string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(
		`UPDATE follow_requests SET status = 'accepted' 
		WHERE requester_id = ? AND target_id = ? AND status = 'pending'`, requesterID, targetID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`INSERT OR IGNORE INTO followers (follower_id, following_id) VALUES (?, ?)`, requesterID, targetID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func DeclineFollowRequest(requesterID, targetID string) error {
	_, err := database.DB.Exec(`
	UPDATE follow_requests SET status = 'declined'
	WHERE requester_id = ? AND target_id = ? AND status = 'pending'`, requesterID, targetID)
	return err
}

func RemoveFollower(followerID, followingID string) error {
	_, err := database.DB.Exec(`DELETE FROM followers WHERE follower_id = ? AND following_id = ?`, followerID, followingID)
	return err
}

func ListFollowers(userID string) ([]string, error) {
	rows, err := database.DB.Query(`
	SELECT follower_id FROM followers WHERE following_id = ?`,
		userID)
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

func ListFollowing(userID string) ([]string, error) {
	rows, err := database.DB.Query(`
	SELECT following_id 
	FROM followers WHERE follower_id = ?`, userID)
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

func ListPendingFollowRequests(targetID string) ([]string, error) {
	rows, err := database.DB.Query(`
	SELECT requester_id 
	FROM follow_requests WHERE target_id = ? AND status = 'pending'`, targetID)
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

func FollowUser(followerID, followingID string) error {
	stmt, err := database.DB.Prepare(`
		INSERT INTO followers (follower_id, following_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(followerID, followingID)
	return err
}
