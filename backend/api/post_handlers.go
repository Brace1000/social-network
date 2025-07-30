package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/database"
	"social-network/database/models"

	"github.com/gorilla/mux"
)

// LikeRequest defines the structure for a like/dislike request body.
type LikeRequest struct {
	LikeType int `json:"like_type"` // 1 for like, -1 for dislike, 0 to remove vote
}

// PostHandlers holds dependencies for post-related handlers.
type PostHandlers struct{}


func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}


func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// NewPostHandlers creates a new PostHandlers.
func NewPostHandlers() *PostHandlers {
	return &PostHandlers{}
}

// LikePostHandler handles liking, disliking, or removing a vote from a post.
func (h *PostHandlers) LikePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate like_type
	if req.LikeType < -1 || req.LikeType > 1 {
		respondWithError(w, http.StatusBadRequest, "Invalid like_type value. Must be -1, 0, or 1.")
		return
	}

	// Check if user has permission to view the post (and thus like it)
	canView, err := CanUserViewPost(userID, postID)
	if err != nil || !canView {
		respondWithError(w, http.StatusForbidden, "You do not have permission to interact with this post")
		return
	}

	// Perform the database operation
	if err := SetPostLike(userID, postID, req.LikeType); err != nil {
		log.Printf("Error setting post like: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process like/dislike")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Vote processed successfully"})
}

// LikeCommentHandler handles liking, disliking, or removing a vote from a comment.
func (h *PostHandlers) LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	commentID, err := strconv.Atoi(vars["commentID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.LikeType < -1 || req.LikeType > 1 {
		respondWithError(w, http.StatusBadRequest, "Invalid like_type value. Must be -1, 0, or 1.")
		return
	}

	// TODO: You might want a CanUserViewComment function for extra security
	// For now, we assume if you can get the comment ID, you can view it.

	if err := SetCommentLike(userID, commentID, req.LikeType); err != nil {
		log.Printf("Error setting comment like: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process like/dislike")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Vote processed successfully"})
}

// SetPostLike inserts, updates, or deletes a user's vote on a post.
func SetPostLike(userID, postID, likeType int) error {
	// If likeType is 0, we delete the vote.
	if likeType == 0 {
		_, err := database.DB.Exec("DELETE FROM post_likes WHERE user_id = ? AND post_id = ?", userID, postID)
		return err
	}

	const query = `
		INSERT INTO post_likes (user_id, post_id, like_type)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, post_id) DO UPDATE SET
			like_type = excluded.like_type;
	`
	_, err := database.DB.Exec(query, userID, postID, likeType)
	return err
}

// SetCommentLike inserts, updates, or deletes a user's vote on a comment.
func SetCommentLike(userID, commentID, likeType int) error {
	if likeType == 0 {
		_, err := database.DB.Exec("DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?", userID, commentID)
		return err
	}
	const query = `
		INSERT INTO comment_likes (user_id, comment_id, like_type)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, comment_id) DO UPDATE SET
			like_type = excluded.like_type;
	`
	_, err := database.DB.Exec(query, userID, commentID, likeType)
	return err
}

// GetFeedForUser retrieves all posts visible to a user, now including like counts.
func GetFeedForUser(userID int) ([]models.PostWithAuthor, error) {
	// This query is now more complex. It uses subqueries to calculate
	// like/dislike counts for each post and to check the current user's reaction.
	const query = `
		SELECT
			p.id, p.user_id, p.content, p.image_url, p.privacy, p.created_at,
			u.first_name, u.last_name, u.nickname, u.avatar_url,
			-- Subquery for like count
			(SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id AND pl.like_type = 1) AS like_count,
			-- Subquery for dislike count
			(SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id AND pl.like_type = -1) AS dislike_count,
			-- Subquery for the current user's vote. COALESCE returns 0 if the user hasn't voted.
			COALESCE((SELECT pl.like_type FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = ?), 0) AS current_user_like_type
		FROM
			posts p
		JOIN
			users u ON p.user_id = u.id
		WHERE
			p.privacy = 'public'
			OR p.user_id = ?
			OR (p.privacy = 'almost_private' AND p.user_id IN (SELECT followed_id FROM followers WHERE follower_id = ?))
			OR (p.privacy = 'private' AND EXISTS (SELECT 1 FROM post_allowed_users pau WHERE pau.post_id = p.id AND pau.user_id = ?))
		ORDER BY
			p.created_at DESC
		LIMIT 50;
	`

	rows, err := database.DB.Query(query, userID, userID, userID, userID, userID)
	if err != nil {
		log.Printf("Error querying user feed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithAuthor
	for rows.Next() {
		var p models.PostWithAuthor
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Content, &p.ImageURL, &p.Privacy, &p.CreatedAt,
			&p.AuthorFirstName, &p.AuthorLastName, &p.AuthorNickname, &p.AuthorAvatarURL,
			&p.LikeCount, &p.DislikeCount, &p.CurrentUserLikeType, // Scan the new fields
		); err != nil {
			log.Printf("Error scanning feed post: %v", err)
			continue
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}


func (h *PostHandlers) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	var req struct {
		Content      string `json:"content"`
		ImageURL     string `json:"image_url"`
		Privacy      string `json:"privacy"`
		AllowedUsers []int  `json:"allowed_users"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Post content cannot be empty")
		return
	}
	if req.Privacy != "public" && req.Privacy != "almost_private" && req.Privacy != "private" {
		respondWithError(w, http.StatusBadRequest, "Invalid privacy setting")
		return
	}
	if req.Privacy == "private" && len(req.AllowedUsers) == 0 {
		respondWithError(w, http.StatusBadRequest, "Private posts must specify at least one allowed user")
		return
	}
	post := models.Post{
		UserID:   userID,
		Content:  req.Content,
		ImageURL: req.ImageURL,
		Privacy:  req.Privacy,
	}
	postID, err := CreatePost(post)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}
	if post.Privacy == "private" {
		if err := AddAllowedUsersForPost(postID, req.AllowedUsers); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to set post permissions")
			return
		}
	}
	post.ID = postID
	respondWithJSON(w, http.StatusCreated, post)
}

func (h *PostHandlers) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}
	canView, err := CanUserViewPost(userID, postID)
	if err != nil || !canView {
		respondWithError(w, http.StatusForbidden, "You do not have permission to comment on this post")
		return
	}
	var req struct {
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Comment content cannot be empty")
		return
	}
	comment := models.Comment{
		PostID:   postID,
		UserID:   userID,
		Content:  req.Content,
		ImageURL: req.ImageURL,
	}
	newComment, err := CreateComment(comment)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}
	respondWithJSON(w, http.StatusCreated, newComment)
}

func (h *PostHandlers) GetFeedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}
	posts, err := GetFeedForUser(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve feed")
		return
	}
	respondWithJSON(w, http.StatusOK, posts)
}

func CreatePost(post models.Post) (int, error) {
	stmt, err := database.DB.Prepare("INSERT INTO posts (user_id, content, image_url, privacy) VALUES (?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(post.UserID, post.Content, post.ImageURL, post.Privacy)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func AddAllowedUsersForPost(postID int, allowedUsers []int) error {
	stmt, err := database.DB.Prepare("INSERT INTO post_allowed_users (post_id, user_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, userID := range allowedUsers {
		_, err := stmt.Exec(postID, userID)
		if err != nil {
			log.Printf("Could not add user %d to post %d: %v", userID, postID, err)
		}
	}
	return nil
}

func CanUserViewPost(userID, postID int) (bool, error) {
	var privacy string
	var authorID int
	err := database.DB.QueryRow("SELECT privacy, user_id FROM posts WHERE id = ?", postID).Scan(&privacy, &authorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Post doesn't exist
		}
		return false, err
	}
	if userID == authorID {
		return true, nil
	}
	switch privacy {
	case "public":
		return true, nil
	case "almost_private":
		// You need a real follower check here
		var isFollowing bool
		err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'accepted')", userID, authorID).Scan(&isFollowing)
		return isFollowing, err
	case "private":
		var count int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM post_allowed_users WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&count)
		return count > 0, err
	default:
		return false, nil
	}
}

func CreateComment(comment models.Comment) (*models.Comment, error) {
	stmt, err := database.DB.Prepare("INSERT INTO comments (post_id, user_id, content, image_url) VALUES (?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(comment.PostID, comment.UserID, comment.Content, comment.ImageURL)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	comment.ID = int(id)
	err = database.DB.QueryRow("SELECT created_at FROM comments WHERE id = ?", id).Scan(&comment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
