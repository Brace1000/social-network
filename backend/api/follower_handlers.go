package api

import (
	"encoding/json"
	"log"
	"net/http"
	"social-network/database/models"
	"social-network/services"

	"github.com/gorilla/mux"
)

// SendFollowRequestHandler handles sending a follow request or auto-following if public.
func (h *UserHandler) SendFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	targetUserID := vars["userId"]
	if targetUserID == "" || targetUserID == actor.ID {
		http.Error(w, "Invalid target user", http.StatusBadRequest)
		return
	}
	targetUser, err := models.GetUserByID(targetUserID)
	if err != nil || targetUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if targetUser.IsPublic {
		err := models.AcceptFollowRequest(actor.ID, targetUserID)
		if err != nil {
			http.Error(w, "Failed to follow user", http.StatusInternalServerError)
			return
		}

		msg := actor.FirstName + " " + actor.LastName + " is now following you."
		notif := &models.Notification{
			UserID:  targetUserID,
			ActorID: actor.ID,
			Type:    "follow_accepted",
			Message: msg,
			Read:    false,
		}
		_ = models.CreateNotification(notif)
		json.NewEncoder(w).Encode(map[string]string{"message": "You are now following this user."})
		return
	}

	err = models.CreateFollowRequest(actor.ID, targetUserID)
	if err != nil {
		http.Error(w, "Failed to send follow request", http.StatusInternalServerError)
		return
	}
	// Notify target user of follow request
	msg := actor.FirstName + " " + actor.LastName + " wants to follow you."
	notif := &models.Notification{
		UserID:  targetUserID,
		ActorID: actor.ID,
		Type:    "follow_request",
		Message: msg,
		Read:    false,
	}
	_ = models.CreateNotification(notif)
	json.NewEncoder(w).Encode(map[string]string{"message": "Follow request sent."})
}

// UnfollowHandler handles unfollowing a user.
func (h *UserHandler) UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	targetUserID := vars["userId"]
	if targetUserID == "" {
		http.Error(w, "Invalid target user", http.StatusBadRequest)
		return
	}
	err := models.RemoveFollower(actor.ID, targetUserID)
	if err != nil {
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Unfollowed user."})
}

// ListFollowersHandler returns a list of followers for a user.
func (h *UserHandler) ListFollowersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	if userID == "" {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	followers, err := models.ListFollowers(userID)
	if err != nil {
		http.Error(w, "Failed to list followers", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"followers": followers})
}

// ListFollowingHandler returns a list of users the user is following.
func (h *UserHandler) ListFollowingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	if userID == "" {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	following, err := models.ListFollowing(userID)
	if err != nil {
		http.Error(w, "Failed to list following", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"following": following})
}

// ListPendingFollowRequestsHandler returns a list of pending follow requests for the current user.
func (h *UserHandler) ListPendingFollowRequestsHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get full follow request objects with all details
	followRequests, err := models.GetPendingFollowRequestsForUser(actor.ID)
	if err != nil {
		http.Error(w, "Failed to list pending follow requests", http.StatusInternalServerError)
		return
	}

	// Get requester details for each request
	var requests []map[string]interface{}
	for _, followRequest := range followRequests {
		requester, err := models.GetUserByID(followRequest.RequesterID)
		if err != nil {
			log.Printf("Error getting requester info for ID %s: %v", followRequest.RequesterID, err)
			continue
		}

		requests = append(requests, map[string]interface{}{
			"id": followRequest.ID,
			"requester": map[string]interface{}{
				"id":         requester.ID,
				"firstName":  requester.FirstName,
				"lastName":   requester.LastName,
				"nickname":   requester.Nickname,
				"avatarPath": requester.AvatarPath,
			},
			"createdAt": followRequest.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests) // Return the array directly, not wrapped
}
