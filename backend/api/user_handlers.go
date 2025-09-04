package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"social-network/database/models"
	"social-network/services"

	"github.com/gorilla/mux"
)

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

	allowed := false
	actor, actorOk := r.Context().Value(services.UserContextKey).(*models.User)
	if targetUser.IsPublic {
		allowed = true
	} else if actorOk && actor.ID == targetUserID {
		allowed = true
	} else if actorOk {
		isFollower, err := models.AreFollowing(actor.ID, targetUserID)
		if err == nil && isFollower {
			allowed = true
		}
	}

	if !allowed {
		http.Error(w, "This profile is private.", http.StatusForbidden)
		return
	}

	followerIDs, errFollowers := models.ListFollowers(targetUserID)
	followers := []*models.User{}
	if errFollowers != nil {
		fmt.Printf("Error fetching followers for user %s: %v", targetUserID, errFollowers)
	} else {
		for _, id := range followerIDs {
			user, err := models.GetUserByID(id)
			if err == nil && user != nil {
				followers = append(followers, user)
			}
		}
	}
	followingIDs, errFollowing := models.ListFollowing(targetUserID)
	following := []*models.User{}
	if errFollowing != nil {
		fmt.Printf("Error fetching following for user %s: %v", targetUserID, errFollowing)
	} else {
		for _, id := range followingIDs {
			user, err := models.GetUserByID(id)
			if err == nil && user != nil {
				following = append(following, user)
			}
		}
	}

	// Only return public info for followers/following
	serializeUser := func(u *models.User) map[string]interface{} {
		return map[string]interface{}{
			"id":         u.ID,
			"firstName":  u.FirstName,
			"lastName":   u.LastName,
			"nickname":   u.Nickname,
			"avatarPath": u.AvatarPath,
			"aboutMe":    u.AboutMe,
			"isPublic":   u.IsPublic,
		}
	}

	resp := map[string]interface{}{
		"id":          targetUser.ID,
		"firstName":   targetUser.FirstName,
		"lastName":    targetUser.LastName,
		"nickname":    targetUser.Nickname,
		"email":       targetUser.Email,
		"dateOfBirth": targetUser.DateOfBirth,
		"avatarPath":  targetUser.AvatarPath,
		"aboutMe":     targetUser.AboutMe,
		"isPublic":    targetUser.IsPublic,
		"followers":   make([]map[string]interface{}, 0),
		"following":   make([]map[string]interface{}, 0),
	}
	for _, u := range followers {
		resp["followers"] = append(resp["followers"].([]map[string]interface{}), serializeUser(u))
	}
	for _, u := range following {
		resp["following"] = append(resp["following"].([]map[string]interface{}), serializeUser(u))
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
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Nickname  string `json:"nickname"`
		AboutMe   string `json:"aboutMe"`
		IsPublic  *bool  `json:"isPublic"`
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

func (h *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := models.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	var filtered []map[string]interface{}
	for _, user := range users {
		if ok && user.ID == actor.ID {
			continue
		}
		isFollowing := false
		if ok {
			isFollowing, _ = models.AreFollowing(actor.ID, user.ID)
		}
		filtered = append(filtered, map[string]interface{}{
			"id":          user.ID,
			"firstName":   user.FirstName,
			"lastName":    user.LastName,
			"nickname":    user.Nickname,
			"email":       user.Email,
			"dateOfBirth": user.DateOfBirth,
			"avatarPath":  user.AvatarPath,
			"aboutMe":     user.AboutMe,
			"isPublic":    user.IsPublic,
			"isFollowing": isFollowing,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

// GetNotificationsHandler returns all notifications for the current user.
func (h *UserHandler) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notifications, err := models.GetNotificationsForUser(actor.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var response []map[string]interface{}
	for _, notif := range notifications {
		notificationData := map[string]interface{}{
			"id":        notif.ID,
			"type":      notif.Type,
			"message":   notif.Message,
			"read":      notif.Read,
			"createdAt": notif.CreatedAt,
		}
		if notif.ActorID != "" {
			actor, err := models.GetUserByID(notif.ActorID)
			if err == nil && actor != nil {
				notificationData["actor"] = map[string]interface{}{
					"id":         actor.ID,
					"firstName":  actor.FirstName,
					"lastName":   actor.LastName,
					"nickname":   actor.Nickname,
					"avatarPath": actor.AvatarPath,
				}
			}
		}

		response = append(response, notificationData)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) MarkNotificationAsReadHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	notificationID, ok := vars["notificationId"]
	if !ok {
		http.Error(w, "Notification ID not provided", http.StatusBadRequest)
		return
	}

	err := models.MarkNotificationAsRead(notificationID, actor.ID)
	if err != nil {
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
}

func (h *UserHandler) ToggleProfilePrivacyHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currentUser, err := models.GetUserByID(actor.ID)
	if err != nil {
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	newPrivacySetting := !currentUser.IsPublic
	err = models.SetUserProfilePrivacy(actor.ID, newPrivacySetting)
	if err != nil {
		http.Error(w, "Failed to update profile privacy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Profile privacy updated successfully",
		"isPublic": newPrivacySetting,
	})
}
