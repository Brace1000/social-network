package api

import (
	"encoding/json"
	"net/http"

	"social-network/database/models"
	"social-network/services"

	"github.com/google/uuid"
	"log"
	"social-network/websocket"
)
var hub *websocket.Hub

// InitUserHandlers initializes the handlers with the hub
func InitUserHandlers(h *websocket.Hub) {
	hub = h
}

// --- Request/Response Structs ---
type RegisterRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Nickname    string `json:"nickname,omitempty"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DateOfBirth string `json:"dateOfBirth"`
	AboutMe     string `json:"aboutMe,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse is a safe structure to send user data to the client (no password hash).
type UserResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname,omitempty"`
}

// --- Handlers ---
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("ERROR: Failed to create user in database: %v", err) // <-- ADD THIS LINE
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	services.ClearSessionCookie(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// CurrentUserHandler gets the currently logged-in user from the context (set by AuthMiddleware).
func CurrentUserHandler(w http.ResponseWriter, r *http.Request) {
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
// FollowRequestHandler handles a user's request to follow another user.
func FollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the current user from the context (set by AuthMiddleware)
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Get the ID of the user to be followed from the URL
	vars := mux.Vars(r)
	targetUserID, ok := vars["userId"]
	if !ok {
		http.Error(w, "User ID not provided", http.StatusBadRequest)
		return
	}

	// 3. Get the target user's profile from the database
	targetUser, err := models.GetUserByID(targetUserID)
	if err != nil || targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 4. Check if the target user has a private profile
	if !targetUser.IsPublic {
		// 5. If private, send a notification
		notificationMessage := fmt.Sprintf("%s %s wants to follow you.", actor.FirstName, actor.LastName)
		
		// Use the hub to send the notification
		go hub.SendNotification(targetUser.ID, actor.ID, "follow_request", notificationMessage)
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Follow request sent."})
	} else {
		// If public, automatically follow them (logic to be added here)
		// For now, just send a response.
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "You are now following " + targetUser.FirstName})
	}
}

