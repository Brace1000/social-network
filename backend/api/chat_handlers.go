package api

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"social-network/database"
	"social-network/database/models"
	"social-network/websocket" 

	"github.com/gorilla/mux"
)

// ChatHandlers holds dependencies for chat-related handlers.
type ChatHandlers struct {
	hub *websocket.Hub
}

// NewChatHandlers creates a new ChatHandlers.

func NewChatHandlers(hub *websocket.Hub) *ChatHandlers {
	return &ChatHandlers{hub: hub}
}

// GetPrivateConversationHandler fetches the message history between the logged-in user and another user.
func (h *ChatHandlers) GetPrivateConversationHandler(w http.ResponseWriter, r *http.Request) {
	
	currentUserID, ok := r.Context().Value(userContextKey).(int)
	if !ok {
		// This should theoretically not happen if AuthMiddleware is working.
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	// Get the other user's ID from the URL path.
	vars := mux.Vars(r)
	otherUserID, err := strconv.Atoi(vars["userID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID provided in URL")
		return
	}

	messages, err := getPrivateMessages(currentUserID, otherUserID)
	if err != nil {
		log.Printf("Error fetching private conversation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve messages")
		return
	}

	respondWithJSON(w, http.StatusOK, messages)
}

// GetGroupConversationHandler fetches the message history for a specific group.
func (h *ChatHandlers) GetGroupConversationHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := r.Context().Value(userContextKey).(int)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	vars := mux.Vars(r)
	// Group IDs are typically strings (UUIDs), so we don't convert to int.
	groupID := vars["groupID"]

	// AUDIT POINT: First, check if the user is a member of this group before fetching messages.
	isMember, err := models.IsUserInGroup(strconv.Itoa(currentUserID), groupID)
	if err != nil {
		log.Printf("Error checking group membership: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not verify group membership")
		return
	}
	if !isMember {
		respondWithError(w, http.StatusForbidden, "Access denied: You are not a member of this group")
		return
	}

	messages, err := getGroupMessages(groupID)
	if err != nil {
		log.Printf("Error fetching group conversation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve group messages")
		return
	}

	respondWithJSON(w, http.StatusOK, messages)
}

// --- Database Helper Functions ---

// getPrivateMessages queries the database for the conversation between two users.
func getPrivateMessages(userID1, userID2 int) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, group_id, content, created_at FROM chat_messages
		WHERE (sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)
		ORDER BY created_at ASC
		LIMIT 100` // Always use LIMIT for chat history to prevent fetching huge datasets.

	rows, err := database.DB.Query(query, userID1, userID2, userID2, userID1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

// getGroupMessages queries the database for all messages in a specific group.
func getGroupMessages(groupID string) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, group_id, content, created_at FROM chat_messages
		WHERE group_id = ?
		ORDER BY created_at ASC
		LIMIT 100`

	rows, err := database.DB.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

// scanMessages is a helper function to reduce code duplication when scanning message rows.
// It correctly handles NULLable fields from the database.
func scanMessages(rows *sql.Rows) ([]models.Message, error) {
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var recipientID, groupID sql.NullString // Use sql.NullString for nullable columns.

		if err := rows.Scan(&msg.ID, &msg.SenderID, &recipientID, &groupID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}

		// Only assign the string if the database value was not NULL.
		if recipientID.Valid {
			msg.RecipientID = recipientID.String
		}
		if groupID.Valid {
			msg.GroupID = groupID.String
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}
