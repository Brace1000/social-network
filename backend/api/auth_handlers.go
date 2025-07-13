package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	ID        string `json:"id"`
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
    // This code remains exactly the same as before.
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if !services.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	sessionToken, err := services.CreateSession(user.ID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	services.SetSessionCookie(w, sessionToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Nickname:  user.Nickname,
	})
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

	targetUser, err := models.GetUserByID(targetUserID)
	if err != nil || targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if !targetUser.IsPublic {
		notificationMessage := fmt.Sprintf("%s %s wants to follow you.", actor.FirstName, actor.LastName)
		go h.hub.SendNotification(targetUser.ID, actor.ID, "follow_request", notificationMessage)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Follow request sent."})
	} else {
		// Automatically follow logic here
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "You are now following " + targetUser.FirstName})
	}
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