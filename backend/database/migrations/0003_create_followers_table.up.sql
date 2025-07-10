-- Up Migration: Creates the followers table
CREATE TABLE IF NOT EXISTS followers (
    follower_id TEXT NOT NULL,          -- The ID of the user who is following
    following_id TEXT NOT NULL,         -- The ID of the user being followed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, following_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE
);