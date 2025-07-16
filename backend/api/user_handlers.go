package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"social-network/database/models"
	"social-network/services"
	"github.com/gorilla/mux"
)

// GetProfileHandler returns a user's profile, respecting privacy settings.
func (h *UserHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
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

	// Default: not allowed to view private profile
	allowed := false
	actor, actorOk := r.Context().Value(services.UserContextKey).(*models.User)
	if targetUser.IsPublic {
		allowed = true
	} else if actorOk && actor.ID == targetUserID {
		allowed = true // owner can view
	} else if actorOk {
		// Check if actor is a follower
		isFollower, err := models.AreFollowing(actor.ID, targetUserID)
		if err == nil && isFollower {
			allowed = true
		}
	}

	if !allowed {
		http.Error(w, "This profile is private.", http.StatusForbidden)
		return
	}

	resp := map[string]interface{}{
		"id": targetUser.ID,
		"firstName": targetUser.FirstName,
		"lastName": targetUser.LastName,
		"nickname": targetUser.Nickname,
		"email": targetUser.Email,
		"dateOfBirth": targetUser.DateOfBirth,
		"avatarPath": targetUser.AvatarPath,
		"aboutMe": targetUser.AboutMe,
		"isPublic": targetUser.IsPublic,
		// TODO: Add posts, followers, following counts if needed
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateProfileHandler allows the authenticated user to update their profile.
func (h *UserHandler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type UpdateProfileRequest struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Nickname    string `json:"nickname"`
		AboutMe     string `json:"aboutMe"`
		IsPublic    *bool  `json:"isPublic"`
	}
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByID(actor.ID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.AboutMe != "" {
		user.AboutMe = req.AboutMe
	}
	if req.IsPublic != nil {
		user.IsPublic = *req.IsPublic
	}

	err = models.UpdateUserProfile(user)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

// UploadAvatarHandler allows the authenticated user to upload/change their avatar.
func (h *UserHandler) UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Could not get avatar file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Only allow JPEG, PNG, GIF
	allowedTypes := map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true}
	if !allowedTypes[handler.Header.Get("Content-Type")] {
		http.Error(w, "Unsupported file type", http.StatusBadRequest)
		return
	}

	// Save file to disk (e.g., ./uploads/avatars/{userID}_{filename})
	avatarPath := fmt.Sprintf("uploads/avatars/%s_%s", actor.ID, handler.Filename)
	out, err := services.SaveUploadedFile(file, avatarPath)
	if err != nil {
		http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Update user record
	user, err := models.GetUserByID(actor.ID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	user.AvatarPath = avatarPath
	if err := models.UpdateUserProfile(user); err != nil {
		http.Error(w, "Failed to update avatar path", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Avatar uploaded successfully", "avatarPath": avatarPath})
} 