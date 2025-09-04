package models

import "time"

type Post struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url,omitempty"`
	Privacy   string    `json:"privacy"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID                  int       `json:"id"`
	PostID              int       `json:"post_id"`
	UserID              string    `json:"user_id"`
	Content             string    `json:"content"`
	ImageURL            string    `json:"image_url,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	LikeCount           int       `json:"like_count"`
	DislikeCount        int       `json:"dislike_count"`
	CurrentUserLikeType int       `json:"current_user_like_type"`
}

type PostWithAuthor struct {
	Post

	AuthorFirstName string `json:"author_first_name"`
	AuthorLastName  string `json:"author_last_name"`
	AuthorNickname  string `json:"author_nickname,omitempty"`
	AuthorAvatarURL string `json:"author_avatar_url,omitempty"`

	LikeCount           int `json:"like_count"`
	DislikeCount        int `json:"dislike_count"`
	CurrentUserLikeType int `json:"current_user_like_type"`
}
