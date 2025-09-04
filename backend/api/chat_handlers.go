package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"social-network/database"
	"social-network/database/models"
	"social-network/services"
	"social-network/websocket"

	"github.com/gorilla/mux"
)

type ChatHandlers struct {
	hub *websocket.Hub
}

func NewChatHandlers(hub *websocket.Hub) *ChatHandlers {
	return &ChatHandlers{hub: hub}
}

func (h *ChatHandlers) GetPrivateConversationHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	vars := mux.Vars(r)
	otherUserID := vars["userID"]
	if otherUserID == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID provided in URL")
		return
	}

	messages, err := getPrivateMessagesWithStrings(currentUser.ID, otherUserID)
	if err != nil {
		log.Printf("Error fetching private conversation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve messages")
		return
	}

	respondWithJSON(w, http.StatusOK, messages)
}

func (h *ChatHandlers) GetGroupConversationHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	vars := mux.Vars(r)
	groupID := vars["groupID"]

	isMember, err := models.IsUserInGroup(currentUser.ID, groupID)
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

func getPrivateMessages(userID1, userID2 int) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, group_id, content, created_at FROM chat_messages
		WHERE (sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)
		ORDER BY created_at ASC
		LIMIT 100`

	rows, err := database.DB.Query(query, userID1, userID2, userID2, userID1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func getPrivateMessagesWithStrings(userID1, userID2 string) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, group_id, content, created_at FROM chat_messages
		WHERE (sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)
		ORDER BY created_at ASC
		LIMIT 100`

	rows, err := database.DB.Query(query, userID1, userID2, userID2, userID1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

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

func scanMessages(rows *sql.Rows) ([]models.Message, error) {
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var recipientID, groupID sql.NullString

		if err := rows.Scan(&msg.ID, &msg.SenderID, &recipientID, &groupID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}

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

func (h *ChatHandlers) GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	conversations, err := getUserConversations(currentUser.ID)
	if err != nil {
		log.Printf("Error fetching conversations: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve conversations")
		return
	}

	respondWithJSON(w, http.StatusOK, conversations)
}

func (h *ChatHandlers) CheckCanMessageHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	vars := mux.Vars(r)
	targetUserID := vars["userID"]
	if targetUserID == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID provided in URL")
		return
	}

	canMessage, err := models.CanUsersMessage(currentUser.ID, targetUserID)
	if err != nil {
		log.Printf("Error checking message permissions: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not verify message permissions")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"canMessage": canMessage})
}

func (h *ChatHandlers) SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	users, err := searchUsersForChat(currentUser.ID, query)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Could not search users")
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (h *ChatHandlers) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Could not identify current user from context")
		return
	}

	var req struct {
		RecipientID string `json:"recipientId"`
		Content     string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RecipientID == "" || req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Recipient ID and content are required")
		return
	}

	canMessage, err := models.CanUsersMessage(currentUser.ID, req.RecipientID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking message permissions")
		return
	}

	if !canMessage {
		respondWithError(w, http.StatusForbidden, "You cannot message this user")
		return
	}

	message := &models.Message{
		SenderID:    currentUser.ID,
		RecipientID: req.RecipientID,
		Content:     req.Content,
		CreatedAt:   time.Now(),
	}

	if err := models.SaveMessage(message); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving message")
		return
	}

	h.hub.SendMessageToUser(req.RecipientID, message)

	respondWithJSON(w, http.StatusCreated, message)
}

type Conversation struct {
	UserID          string `json:"userId,omitempty"`
	GroupID         string `json:"groupId,omitempty"`
	Name            string `json:"name"`
	AvatarPath      string `json:"avatarPath,omitempty"`
	LastMessage     string `json:"lastMessage"`
	LastMessageTime string `json:"lastMessageTime"`
	UnreadCount     int    `json:"unreadCount"`
	Type            string `json:"type"`
}

func getUserConversations(userID string) ([]Conversation, error) {
	var conversations []Conversation

	privateQuery := `
		SELECT DISTINCT
			CASE
				WHEN cm.sender_id = ? THEN cm.recipient_id
				ELSE cm.sender_id
			END as other_user_id,
			u.first_name || ' ' || u.last_name as name,
			u.avatar_path,
			cm.content as last_message,
			cm.created_at as last_message_time
		FROM chat_messages cm
		JOIN users u ON (
			CASE
				WHEN cm.sender_id = ? THEN u.id = cm.recipient_id
				ELSE u.id = cm.sender_id
			END
		)
		WHERE (cm.sender_id = ? OR cm.recipient_id = ?)
			AND cm.recipient_id IS NOT NULL
		ORDER BY cm.created_at DESC
	`

	rows, err := database.DB.Query(privateQuery, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seenUsers := make(map[string]bool)
	for rows.Next() {
		var conv Conversation
		var avatarPath sql.NullString
		var lastMessageTime string

		err := rows.Scan(&conv.UserID, &conv.Name, &avatarPath, &conv.LastMessage, &lastMessageTime)
		if err != nil {
			return nil, err
		}

		if seenUsers[conv.UserID] {
			continue
		}
		seenUsers[conv.UserID] = true

		conv.Type = "private"
		conv.AvatarPath = avatarPath.String
		conv.LastMessageTime = lastMessageTime
		conv.UnreadCount = 0

		conversations = append(conversations, conv)
	}

	groupQuery := `
		SELECT DISTINCT
			g.id as group_id,
			g.name,
			cm.content as last_message,
			cm.created_at as last_message_time
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		LEFT JOIN chat_messages cm ON g.id = cm.group_id
		WHERE gm.user_id = ?
		ORDER BY cm.created_at DESC
	`

	rows, err = database.DB.Query(groupQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seenGroups := make(map[string]bool)
	for rows.Next() {
		var conv Conversation
		var lastMessage, lastMessageTime sql.NullString

		err := rows.Scan(&conv.GroupID, &conv.Name, &lastMessage, &lastMessageTime)
		if err != nil {
			return nil, err
		}

		if seenGroups[conv.GroupID] {
			continue
		}
		seenGroups[conv.GroupID] = true

		conv.Type = "group"
		conv.LastMessage = lastMessage.String
		conv.LastMessageTime = lastMessageTime.String
		conv.UnreadCount = 0

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func searchUsersForChat(currentUserID, query string) ([]map[string]interface{}, error) {
	searchQuery := `
		SELECT id, first_name, last_name, nickname, avatar_path, is_public
		FROM users
		WHERE id != ?
			AND (first_name LIKE ? OR last_name LIKE ? OR nickname LIKE ?)
		LIMIT 20
	`

	searchTerm := "%" + query + "%"
	rows, err := database.DB.Query(searchQuery, currentUserID, searchTerm, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, firstName, lastName string
		var nickname, avatarPath sql.NullString
		var isPublic bool

		err := rows.Scan(&id, &firstName, &lastName, &nickname, &avatarPath, &isPublic)
		if err != nil {
			return nil, err
		}

		canMessage, err := models.CanUsersMessage(currentUserID, id)
		if err != nil {
			log.Printf("Error checking message permissions for user %s: %v", id, err)
			continue
		}

		if !canMessage {
			continue
		}

		user := map[string]interface{}{
			"id":         id,
			"firstName":  firstName,
			"lastName":   lastName,
			"nickname":   nickname.String,
			"avatarPath": avatarPath.String,
			"isPublic":   isPublic,
			"canMessage": canMessage,
		}

		users = append(users, user)
	}

	return users, nil
}
