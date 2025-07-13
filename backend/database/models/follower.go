package models

import (
	"social-network/database"
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
