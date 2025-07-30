package models

import "time"

// Post represents the core data of a single post in the database.
type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"` // This is the author's ID
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url,omitempty"`
	Privacy   string    `json:"privacy"` // 'public', 'almost_private', 'private'
	CreatedAt time.Time `json:"created_at"`
}

// Comment represents a comment on a post.
type Comment struct {
	ID                int       `json:"id"`
	PostID            int       `json:"post_id"`
	UserID            int       `json:"user_id"`
	Content           string    `json:"content"`
	ImageURL          string    `json:"image_url,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	LikeCount         int       `json:"like_count"`
	DislikeCount      int       `json:"dislike_count"`
	CurrentUserLikeType int     `json:"current_user_like_type"` // 1 for like, -1 for dislike, 0 for none
}

// PostWithAuthor is a special struct used for sending feed data to the frontend.
// It combines the post information with the author's public information and like counts.
type PostWithAuthor struct {
	Post // Embeds all fields from the Post struct (ID, Content, etc.)

	AuthorFirstName     string `json:"author_first_name"`
	AuthorLastName      string `json:"author_last_name"`
	AuthorNickname      string `json:"author_nickname,omitempty"`
	AuthorAvatarURL     string `json:"author_avatar_url,omitempty"`

	LikeCount           int `json:"like_count"`
	DislikeCount        int `json:"dislike_count"`
	CurrentUserLikeType int `json:"current_user_like_type"` // 1 for like, -1 for dislike, 0 for none
}