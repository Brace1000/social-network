package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"social-network/database"
	"social-network/database/models"
	"social-network/services"
	"social-network/websocket"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UserHandler holds dependencies for user-related handlers, like the WebSocket hub.
type UserHandler struct {
	hub *websocket.Hub
}

// NewUserHandlers creates a new UserHandler with its dependencies.
func NewUserHandlers(h *websocket.Hub) *UserHandler {
	return &UserHandler{hub: h}
}

// --- Request/Response Structs remain the same ---
type RegisterRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname,omitempty"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DateOfBirth string `json:"dateOfBirth"`
	AboutMe     string `json:"aboutMe,omitempty"`
}

// ... (LoginRequest and UserResponse are the same as before)
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        string `json:"id"` // <-- CHANGED: ID is a STRING
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname,omitempty"`
}

// --- Handlers are now methods on the UserHandler struct ---

func (h *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// This code remains exactly the same as before.
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	existingUser, _ := models.GetUserByEmail(req.Email)
	if existingUser != nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Server error during password hashing", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		ID:           uuid.NewString(),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Nickname:     req.Nickname,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DateOfBirth:  req.DateOfBirth,
		AboutMe:      req.AboutMe,
		IsPublic:     true,
	}

	if err := models.CreateUser(user); err != nil {
		log.Printf("ERROR: Failed to create user in database: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}



func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- LoginHandler: Running ---")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var userID, userEmail, userFirstName, userLastName, userNickname, hashedPassword string
	query := "SELECT id, email, first_name, last_name, nickname, password_hash FROM users WHERE email = ?"
	log.Printf("LoginHandler: Executing DB query to find user: %s", req.Email)
	err := database.DB.QueryRow(query, req.Email).Scan(&userID, &userEmail, &userFirstName, &userLastName, &userNickname, &hashedPassword)
	if err != nil {
		log.Printf("LoginHandler: FAILED finding user in DB. Error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	log.Printf("LoginHandler: SUCCESS finding user in DB. User ID: %s", userID)

	if !services.CheckPasswordHash(req.Password, hashedPassword) {
		log.Println("LoginHandler: FAILED password check.")
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	log.Println("LoginHandler: SUCCESS password check.")

	sessionToken, expiryTime, err := createAndSaveSession(userID)
	if err != nil {
		log.Printf("LoginHandler: FAILED creating session. Error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	log.Printf("LoginHandler: Setting cookie with name 'social_network_session' and value '%s'", sessionToken)
	http.SetCookie(w, &http.Cookie{
		Name:     "social_network_session",
		Value:    sessionToken,
		Expires:  expiryTime,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	log.Println("LoginHandler: Login successful. Sending response.")
	respondWithJSON(w, http.StatusOK, UserResponse{
		ID:        userID,
		FirstName: userFirstName,
		LastName:  userLastName,
		Email:     userEmail,
		Nickname:  userNickname,
	})
}

func createAndSaveSession(userID string) (string, time.Time, error) {
	tokenUUID, err := uuid.NewRandom()
	if err != nil { return "", time.Time{}, err }
	token := tokenUUID.String()
	expiry := time.Now().Add(7 * 24 * time.Hour)

	log.Printf("createAndSaveSession: Attempting to INSERT token '%s' for user '%s' into DB.", token, userID)
	query := "INSERT INTO sessions (token, user_id, expiry) VALUES (?, ?, ?)"
	_, err = database.DB.Exec(query, token, userID, expiry)
	if err != nil {
		log.Printf("createAndSaveSession: FAILED to insert session into DB. Error: %v", err)
		return "", time.Time{}, err
	}

	log.Println("createAndSaveSession: SUCCESS inserting session into DB.")
	return token, expiry, nil
}

func (h *UserHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	services.ClearSessionCookie(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *UserHandler) CurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Nickname:  user.Nickname,
	})
}

func (h *UserHandler) FollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	targetUserID, ok := vars["userId"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	// Prevent self-following
	if actor.ID == targetUserID {
		http.Error(w, "Cannot follow yourself", http.StatusBadRequest)
		return
	}

	targetUser, err := models.GetUserByID(targetUserID)
	if err != nil || targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if already following
	alreadyFollowing, err := models.AreFollowing(actor.ID, targetUserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if alreadyFollowing {
		http.Error(w, "Already following this user", http.StatusBadRequest)
		return
	}

	// Check if there's already a pending request
	existingRequest, err := models.GetFollowRequest(actor.ID, targetUserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingRequest != nil {
		http.Error(w, "Follow request already sent", http.StatusBadRequest)
		return
	}

	if targetUser.IsPublic {
		// For public profiles, automatically follow
		err = models.FollowUser(actor.ID, targetUserID)
		if err != nil {
			http.Error(w, "Failed to follow user", http.StatusInternalServerError)
			return
		}

		// Send real-time updates to both users to refresh their user lists
		go h.hub.SendUserListUpdate(actor.ID)
		go h.hub.SendUserListUpdate(targetUserID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "You are now following " + targetUser.FirstName})
	} else {
		// For private profiles, create a follow request
		err = models.CreateFollowRequest(actor.ID, targetUserID)
		if err != nil {
			http.Error(w, "Failed to create follow request", http.StatusInternalServerError)
			return
		}

		// Send notification to the target user
		notificationMessage := fmt.Sprintf("%s %s wants to follow you.", actor.FirstName, actor.LastName)
		go h.hub.SendNotification(targetUser.ID, actor.ID, "follow_request", notificationMessage)

		// Send follow request updates to both users to refresh their frontend
		go h.hub.SendFollowRequestUpdate(targetUser.ID) // Recipient
		go h.hub.SendFollowRequestUpdate(actor.ID)      // Requester

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Follow request sent."})
	}
}

// AcceptFollowRequestHandler handles accepting a follow request.
func (h *UserHandler) AcceptFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	requestID, ok := vars["requestId"]
	if !ok {
		http.Error(w, "Request ID not provided", http.StatusBadRequest)
		return
	}

	// Get the follow request
	requests, err := models.GetPendingFollowRequestsForUser(actor.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var targetRequest *models.FollowRequest
	for _, req := range requests {
		if req.ID == requestID {
			targetRequest = req
			break
		}
	}

	if targetRequest == nil {
		http.Error(w, "Follow request not found", http.StatusNotFound)
		return
	}

	// Update the request status to accepted
	err = models.UpdateFollowRequestStatus(requestID, "accepted")
	if err != nil {
		http.Error(w, "Failed to accept request", http.StatusInternalServerError)
		return
	}

	// Create the follow relationship
	err = models.FollowUser(targetRequest.RequesterID, targetRequest.TargetID)
	if err != nil {
		http.Error(w, "Failed to create follow relationship", http.StatusInternalServerError)
		return
	}

	// Get requester info for notification
	requester, err := models.GetUserByID(targetRequest.RequesterID)
	if err != nil {
		log.Printf("Error getting requester info: %v", err)
	} else {
		// Send notification to requester that their request was accepted
		notificationMessage := fmt.Sprintf("%s %s accepted your follow request.", actor.FirstName, actor.LastName)
		go h.hub.SendNotification(requester.ID, actor.ID, "follow_accepted", notificationMessage)
	}

	// Send real-time updates to both users to refresh their follow request lists
	go h.hub.SendFollowRequestUpdate(actor.ID)     // Recipient (current user)
	go h.hub.SendFollowRequestUpdate(requester.ID) // Requester

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Follow request accepted"})
}

// DeclineFollowRequestHandler handles declining a follow request.
func (h *UserHandler) DeclineFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	requestID, ok := vars["requestId"]
	if !ok {
		http.Error(w, "Request ID not provided", http.StatusBadRequest)
		return
	}

	// Get the follow request
	requests, err := models.GetPendingFollowRequestsForUser(actor.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var targetRequest *models.FollowRequest
	for _, req := range requests {
		if req.ID == requestID {
			targetRequest = req
			break
		}
	}

	if targetRequest == nil {
		http.Error(w, "Follow request not found", http.StatusNotFound)
		return
	}

	// Update the request status to declined
	err = models.UpdateFollowRequestStatus(requestID, "declined")
	if err != nil {
		http.Error(w, "Failed to decline request", http.StatusInternalServerError)
		return
	}

	// Get requester info for notification
	requester, err := models.GetUserByID(targetRequest.RequesterID)
	if err != nil {
		log.Printf("Error getting requester info: %v", err)
	} else {
		// Send notification to requester that their request was declined
		notificationMessage := fmt.Sprintf("%s %s declined your follow request.", actor.FirstName, actor.LastName)
		go h.hub.SendNotification(requester.ID, actor.ID, "follow_declined", notificationMessage)
	}

	// Send real-time updates to both users to refresh their follow request lists
	go h.hub.SendFollowRequestUpdate(actor.ID)     // Recipient (current user)
	go h.hub.SendFollowRequestUpdate(requester.ID) // Requester

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Follow request declined"})
}

// CancelFollowRequestHandler handles canceling a follow request (for the requester).
func (h *UserHandler) CancelFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	requestID, ok := vars["requestId"]
	if !ok {
		http.Error(w, "Request ID not provided", http.StatusBadRequest)
		return
	}

	// Get the follow request to verify ownership
	request, err := models.GetFollowRequestByID(requestID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if request == nil {
		http.Error(w, "Follow request not found", http.StatusNotFound)
		return
	}

	// Verify that the current user is the requester
	if request.RequesterID != actor.ID {
		http.Error(w, "Unauthorized to cancel this request", http.StatusForbidden)
		return
	}

	// Delete the follow request
	err = models.DeleteFollowRequest(requestID)
	if err != nil {
		http.Error(w, "Failed to cancel request", http.StatusInternalServerError)
		return
	}

	// Get recipient info for notification
	recipient, err := models.GetUserByID(request.TargetID)
	if err != nil {
		log.Printf("Error getting recipient info: %v", err)
	} else {
		// Send notification to recipient that the request was canceled
		notificationMessage := fmt.Sprintf("%s %s canceled their follow request.", actor.FirstName, actor.LastName)
		go h.hub.SendNotification(recipient.ID, actor.ID, "follow_canceled", notificationMessage)
	}

	// Send real-time updates to both users to refresh their follow request lists
	go h.hub.SendFollowRequestUpdate(actor.ID)     // Requester (current user)
	go h.hub.SendFollowRequestUpdate(recipient.ID) // Recipient

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Follow request canceled"})
}

// GetFollowRequestsHandler returns all pending follow requests for the current user.
func (h *UserHandler) GetFollowRequestsHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := models.GetPendingFollowRequestsForUser(actor.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get requester details for each request
	var response []map[string]interface{}
	for _, req := range requests {
		requester, err := models.GetUserByID(req.RequesterID)
		if err != nil {
			log.Printf("Error getting requester info for request %s: %v", req.ID, err)
			continue
		}

		response = append(response, map[string]interface{}{
			"id": req.ID,
			"requester": map[string]interface{}{
				"id":         requester.ID,
				"firstName":  requester.FirstName,
				"lastName":   requester.LastName,
				"nickname":   requester.Nickname,
				"avatarPath": requester.AvatarPath,
			},
			"createdAt": req.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMyFollowRequestsHandler returns all pending follow requests sent by the current user.
func (h *UserHandler) GetMyFollowRequestsHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := models.GetPendingFollowRequestsByUser(actor.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get recipient details for each request
	var response []map[string]interface{}
	for _, req := range requests {
		recipient, err := models.GetUserByID(req.TargetID)
		if err != nil {
			log.Printf("Error getting recipient info for request %s: %v", req.ID, err)
			continue
		}

		response = append(response, map[string]interface{}{
			"id": req.ID,
			"recipient": map[string]interface{}{
				"id":         recipient.ID,
				"firstName":  recipient.FirstName,
				"lastName":   recipient.LastName,
				"nickname":   recipient.Nickname,
				"avatarPath": recipient.AvatarPath,
			},
			"createdAt": req.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MakeProfilePrivateHandler is a TEMPORARY handler for debugging purposes.
// It changes a user's profile to private.
func (h *UserHandler) MakeProfilePrivateHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the user ID from the URL path
	vars := mux.Vars(r)
	targetUserID, ok := vars["userId"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	// Perform the database update
	err := models.SetUserProfilePrivacy(targetUserID, false) // false means private
	if err != nil {
		log.Printf("Error updating user privacy: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s profile has been set to PRIVATE for testing.", targetUserID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User %s profile is now private", targetUserID),
	})
}

// CheckFollowRequestStatusHandler checks if there's a pending follow request from actor to target user
func (h *UserHandler) CheckFollowRequestStatusHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	targetUserID, ok := vars["userId"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	// Check if already following
	alreadyFollowing, err := models.AreFollowing(actor.ID, targetUserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if there's a pending follow request
	existingRequest, err := models.GetFollowRequest(actor.ID, targetUserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	status := "not_following"
	if alreadyFollowing {
		status = "following"
	} else if existingRequest != nil && existingRequest.Status == "pending" {
		status = "request_sent"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": status,
	})
}
