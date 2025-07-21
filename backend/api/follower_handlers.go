package api

import (
	"encoding/json"
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
		// Auto-follow
		err := models.AcceptFollowRequest(actor.ID, targetUserID)
		if err != nil {
			http.Error(w, "Failed to follow user", http.StatusInternalServerError)
			return
		}
		// Notify target user of new follower
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
	// Create follow request
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

// AcceptFollowRequestHandler handles accepting a follow request.
func (h *UserHandler) AcceptFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	requesterID := vars["userId"]
	if requesterID == "" {
		http.Error(w, "Invalid requester user", http.StatusBadRequest)
		return
	}
	err := models.AcceptFollowRequest(requesterID, actor.ID)
	if err != nil {
		http.Error(w, "Failed to accept follow request", http.StatusInternalServerError)
		return
	}
	// Notify requester that their request was accepted
	msg := actor.FirstName + " " + actor.LastName + " accepted your follow request."
	notif := &models.Notification{
		UserID:  requesterID,
		ActorID: actor.ID,
		Type:    "follow_accepted",
		Message: msg,
		Read:    false,
	}
	_ = models.CreateNotification(notif)
	json.NewEncoder(w).Encode(map[string]string{"message": "Follow request accepted."})
}

// DeclineFollowRequestHandler handles declining a follow request.
func (h *UserHandler) DeclineFollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	requesterID := vars["userId"]
	if requesterID == "" {
		http.Error(w, "Invalid requester user", http.StatusBadRequest)
		return
	}
	err := models.DeclineFollowRequest(requesterID, actor.ID)
	if err != nil {
		http.Error(w, "Failed to decline follow request", http.StatusInternalServerError)
		return
	}
	// Optionally notify requester of decline
	msg := actor.FirstName + " " + actor.LastName + " declined your follow request."
	notif := &models.Notification{
		UserID:  requesterID,
		ActorID: actor.ID,
		Type:    "follow_declined",
		Message: msg,
		Read:    false,
	}
	_ = models.CreateNotification(notif)
	json.NewEncoder(w).Encode(map[string]string{"message": "Follow request declined."})
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
	json.NewEncoder(w).Encode(map[string]interface{}{ "followers": followers })
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
	json.NewEncoder(w).Encode(map[string]interface{}{ "following": following })
}

// ListPendingFollowRequestsHandler returns a list of pending follow requests for the current user.
func (h *UserHandler) ListPendingFollowRequestsHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(services.UserContextKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requests, err := models.ListPendingFollowRequests(actor.ID)
	if err != nil {
		http.Error(w, "Failed to list pending follow requests", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{ "pendingRequests": requests })
} 