package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/database" // You will need to implement these functions
	"social-network/database/models"

	"github.com/gorilla/mux"
)

// PostHandlers holds dependencies for post-related handlers.
type PostHandlers struct {
	// db *database.DB // Example of a DB connection dependency
}

// respondWithError sends a JSON error response with the given status code and message.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response with the given status code and payload.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// NewPostHandlers creates a new PostHandlers.
func NewPostHandlers() *PostHandlers {
	return &PostHandlers{}
}

// CreatePostHandler handles the creation of a new post.
func (h *PostHandlers) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get User ID from authentication middleware
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// 2. Decode the request body
	var req struct {
		Content      string `json:"content"`
		ImageURL     string `json:"image_url"`
		Privacy      string `json:"privacy"`
		AllowedUsers []int  `json:"allowed_users"` // For 'private' posts
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 3. Validate input
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

	// 4. Create the post in the database
	post := models.Post{
		UserID:   userID,
		Content:  req.Content,
		ImageURL: req.ImageURL,
		Privacy:  req.Privacy,
	}

	// Assume database.CreatePost returns the ID of the new post
	postID, err := CreatePost(post)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	// 5. If privacy is 'private', link allowed users
	if post.Privacy == "private" {
		if err := AddAllowedUsersForPost(postID, req.AllowedUsers); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to set post permissions")
			return
		}
	}

	post.ID = postID
	respondWithJSON(w, http.StatusCreated, post)
}

// CreateCommentHandler handles adding a comment to a post.
func (h *PostHandlers) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get User ID from middleware and Post ID from URL
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

	// 2. IMPORTANT: Check if the user has permission to view (and thus comment on) the post
	canView, err := CanUserViewPost(userID, postID)
	if err != nil || !canView {
		respondWithError(w, http.StatusForbidden, "You do not have permission to comment on this post")
		return
	}

	// 3. Decode request body
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

	// 4. Insert the comment into the database
	comment := models.Comment{
		PostID:   postID,
		UserID:   userID,
		Content:  req.Content,
		ImageURL: req.ImageURL,
	}

	// Assume database.CreateComment returns the new comment with its ID
	newComment, err := CreateComment(comment)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	respondWithJSON(w, http.StatusCreated, newComment)
}

// GetFeedPostsHandler gets all posts visible to the current user.
func (h *PostHandlers) GetFeedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// This database function should contain the complex logic to fetch:
	// - All 'public' posts
	// - 'almost_private' posts from users the current user follows
	// - 'private' posts where the current user is in the post_allowed_users list
	posts, err := GetFeedForUser(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve feed")
		return
	}

	respondWithJSON(w, http.StatusOK, posts)
}

// CreatePost inserts a new post into the database and returns its ID.
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

// AddAllowedUsersForPost links a private post to the users who are allowed to see it.
func AddAllowedUsersForPost(postID int, allowedUsers []int) error {
	stmt, err := database.DB.Prepare("INSERT INTO post_allowed_users (post_id, user_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, userID := range allowedUsers {
		_, err := stmt.Exec(postID, userID)
		if err != nil {
			// Continue trying to insert others, but log the error
			log.Printf("Could not add user %d to post %d: %v", userID, postID, err)
		}
	}
	return nil
}

// CanUserViewPost checks if a user has permission to see a specific post.
func CanUserViewPost(userID, postID int) (bool, error) {
	var privacy string
	var authorID int
	err := database.DB.QueryRow("SELECT privacy, user_id FROM posts WHERE id = ?", postID).Scan(&privacy, &authorID)
	if err != nil {
		return false, err
	}

	if userID == authorID {
		return true, nil
	}

	switch privacy {
	case "public":
		return true, nil
	case "almost_private":
		// You need a function to check for follower status
		// isFollowing, err := IsFollowing(userID, authorID)
		// return isFollowing, err
		return true, nil // Placeholder: You need to implement follower logic here
	case "private":
		var count int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM post_allowed_users WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&count)
		return count > 0, err
	default:
		return false, nil
	}
}

// CreateComment inserts a new comment and returns the full comment model.
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

	// To get the timestamp, we need to query it back
	err = database.DB.QueryRow("SELECT created_at FROM comments WHERE id = ?", id).Scan(&comment.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

// GetFeedForUser retrieves all posts that should be visible to a given user.
// This includes:
// 1. All 'public' posts.
// 2. 'almost_private' posts from users the current user follows.
// 3. 'private' posts where the current user has been explicitly granted access.
// It returns a slice of PostWithAuthor, joining post data with author details.
func GetFeedForUser(userID int) ([]models.PostWithAuthor, error) {
	const query = `
		SELECT
			p.id, p.user_id, p.content, p.image_url, p.privacy, p.created_at,
			u.first_name, u.last_name, u.nickname, u.avatar_url
		FROM
			posts p
		JOIN
			users u ON p.user_id = u.id
		WHERE
			-- Condition 1: The post is public
			p.privacy = 'public'
		OR
			-- Condition 2: The user is the author of the post
			p.user_id = ?
		OR
			-- Condition 3: Post is 'almost_private' and the author is followed by the user
			(p.privacy = 'almost_private' AND p.user_id IN (
				SELECT followed_id FROM followers WHERE follower_id = ?
			))
		OR
			-- Condition 4: Post is 'private' and the user is on the allowed list
			(p.privacy = 'private' AND EXISTS (
				SELECT 1 FROM post_allowed_users pau WHERE pau.post_id = p.id AND pau.user_id = ?
			))
		ORDER BY
			p.created_at DESC
		LIMIT 50; -- Add a limit to avoid overwhelming the client
	`

	// The userID is passed three times to satisfy the three placeholders (?) in the query.
	rows, err := database.DB.Query(query, userID, userID, userID)
	if err != nil {
		log.Printf("Error querying user feed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithAuthor

	for rows.Next() {
		var p models.PostWithAuthor
		// The order of scanning must exactly match the order of columns in the SELECT statement.
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Content, &p.ImageURL, &p.Privacy, &p.CreatedAt,
			&p.AuthorFirstName, &p.AuthorLastName, &p.AuthorNickname, &p.AuthorAvatarURL,
		); err != nil {
			log.Printf("Error scanning feed post: %v", err)
			continue // Skip this post and continue with the next
		}
		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating over feed rows: %v", err)
		return nil, err
	}

	return posts, nil
}
