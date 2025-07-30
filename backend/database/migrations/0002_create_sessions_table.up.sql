CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
   
    user_id TEXT NOT NULL, 
    expiry TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);