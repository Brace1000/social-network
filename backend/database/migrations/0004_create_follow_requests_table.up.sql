-- Up Migration: Creates the follow_requests table
CREATE TABLE IF NOT EXISTS follow_requests (
    id TEXT PRIMARY KEY,
    requester_id TEXT NOT NULL,   -- The user who wants to follow
    target_id TEXT NOT NULL,      -- The user to be followed
    status TEXT NOT NULL CHECK(status IN ('pending', 'accepted', 'declined')) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES users(id) ON DELETE CASCADE
); 