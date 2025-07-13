-- Up Migration: Creates the notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,                      -- The user who receives the notification
    actor_id TEXT,                              -- The user who performed the action (e.g., sent the follow request)
    type TEXT NOT NULL,                         -- e.g., 'follow_request', 'group_invite'
    message TEXT NOT NULL,                      -- The notification text, e.g., "John Doe wants to follow you."
    read INTEGER NOT NULL DEFAULT 0,            -- 0 for unread, 1 for read
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE CASCADE
);