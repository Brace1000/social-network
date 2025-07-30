package models

import (
	"database/sql"
	"social-network/database"
	"time"

	"github.com/google/uuid"
)

// FollowRequest represents the structure of the 'follow_requests' table.
type FollowRequest struct {
	ID          string
	RequesterID string
	TargetID    string
	Status      string // 'pending', 'accepted', 'declined'
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateFollowRequest creates a new follow request in the database.
func CreateFollowRequest(requesterID, targetID string) error {
	requestID := uuid.NewString()
	stmt, err := database.DB.Prepare(`
		INSERT INTO follow_requests (id, requester_id, target_id, status)
		VALUES (?, ?, ?, 'pending')
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(requestID, requesterID, targetID)
	return err
}

// GetFollowRequest retrieves a follow request by requester and target IDs.
func GetFollowRequest(requesterID, targetID string) (*FollowRequest, error) {
	request := &FollowRequest{}
	row := database.DB.QueryRow(`
		SELECT id, requester_id, target_id, status, created_at, updated_at 
		FROM follow_requests 
		WHERE requester_id = ? AND target_id = ?
	`, requesterID, targetID)

	err := row.Scan(
		&request.ID,
		&request.RequesterID,
		&request.TargetID,
		&request.Status,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No request found
		}
		return nil, err
	}

	return request, nil
}

// GetFollowRequestByID retrieves a follow request by its ID.
func GetFollowRequestByID(requestID string) (*FollowRequest, error) {
	request := &FollowRequest{}
	row := database.DB.QueryRow(`
		SELECT id, requester_id, target_id, status, created_at, updated_at 
		FROM follow_requests 
		WHERE id = ?
	`, requestID)

	err := row.Scan(
		&request.ID,
		&request.RequesterID,
		&request.TargetID,
		&request.Status,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No request found
		}
		return nil, err
	}

	return request, nil
}

// GetPendingFollowRequestsForUser retrieves all pending follow requests for a user.
func GetPendingFollowRequestsForUser(userID string) ([]*FollowRequest, error) {
	rows, err := database.DB.Query(`
		SELECT id, requester_id, target_id, status, created_at, updated_at 
		FROM follow_requests 
		WHERE target_id = ? AND status = 'pending'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*FollowRequest
	for rows.Next() {
		request := &FollowRequest{}
		err := rows.Scan(
			&request.ID,
			&request.RequesterID,
			&request.TargetID,
		&request.Status,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// GetPendingFollowRequestsByUser retrieves all pending follow requests sent by a user.
func GetPendingFollowRequestsByUser(userID string) ([]*FollowRequest, error) {
	rows, err := database.DB.Query(`
		SELECT id, requester_id, target_id, status, created_at, updated_at 
		FROM follow_requests 
		WHERE requester_id = ? AND status = 'pending'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*FollowRequest
	for rows.Next() {
		request := &FollowRequest{}
		err := rows.Scan(
			&request.ID,
			&request.RequesterID,
			&request.TargetID,
			&request.Status,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// UpdateFollowRequestStatus updates the status of a follow request.
func UpdateFollowRequestStatus(requestID, status string) error {
	stmt, err := database.DB.Prepare(`
		UPDATE follow_requests 
		SET status = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, requestID)
	return err
}

// DeleteFollowRequest deletes a follow request.
func DeleteFollowRequest(requestID string) error {
	stmt, err := database.DB.Prepare("DELETE FROM follow_requests WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(requestID)
	return err
}

 