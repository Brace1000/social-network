package models

import "time"

type GroupMember struct {
    ID        string    `json:"id"`
    GroupID   string    `json:"group_id"`
    UserID    string    `json:"user_id"`
    Role      string    `json:"role"`
    JoinedAt  time.Time `json:"joined_at"`
}