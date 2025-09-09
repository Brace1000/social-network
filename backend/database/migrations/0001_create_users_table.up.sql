CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    nickname TEXT,
    email TEXT NOT NULL UNIQUE,
   
    password_hash TEXT NOT NULL, 
    date_of_birth TEXT,
    avatar_path TEXT,
    about_me TEXT,
    is_public TEXT DEFAULT 'public',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);