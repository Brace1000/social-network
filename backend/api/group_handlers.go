package api

import (
	"encoding/json"
	"net/http"

	// database "social-network/database"
	"social-network/database/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// CreateGroupHandler creates a new group.
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group.ID = uuid.New().String()

	err = database.CreateGroup(group.ID, group.Title, group.Description, group.CreatorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// GetGroupHandler retrieves a group by ID.
func GetGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	group, err := database.GetGroupByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(group)
}

// UpdateGroupHandler updates a group.
func UpdateGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var group models.Group
	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = database.UpdateGroup(id, group.Title, group.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteGroupHandler deletes a group.
func DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := database.DeleteGroup(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// AddGroupMemberHandler adds a member to a group.
func AddGroupMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupID"]
	userID := vars["userID"]

	id := uuid.New().String()

	err := database.AddGroupMember(id, groupID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveGroupMemberHandler removes a member from a group.
func RemoveGroupMemberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupID"]
	userID := vars["userID"]

	err := database.RemoveGroupMember(groupID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetGroupMembersHandler retrieves all members of a group.
func GetGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupID"]

	members, err := database.GetGroupMembers(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(members)
}

// InviteUserToGroupHandler invites a user to a group.
func InviteUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupID"]
	inviterID := vars["inviterID"]
	inviteeID := vars["inviteeID"]

	id := uuid.New().String()

	err := database.InviteUserToGroup(id, groupID, inviterID, inviteeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// AcceptGroupInvitationHandler accepts a group invitation.
func AcceptGroupInvitationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupID"]
	inviteeID := vars["inviteeID"]

	err := database.AcceptGroupInvitation(groupID, inviteeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RejectGroupInvitationHandler rejects a group invitation.
func RejectGroupInvitationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["groupID"]
	inviteeID := vars["inviteeID"]

	err := database.RejectGroupInvitation(groupID, inviteeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetAllGroupsHandler retrieves all groups.
func GetAllGroupsHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := database.GetAllGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(groups)
}
