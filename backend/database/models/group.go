package models

import "time"

type Group struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    CreatorID   string    `json:"creator_id"`
    CreatedAt   time.Time `json:"created_at"`
}