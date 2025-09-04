CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    nickname TEXT,
    email TEXT NOT NULL UNIQUE,
   
    password_hash TEXT NOT NULL, 
    date_of_birth TEXT,
    about_me TEXT,
    avatar_path TEXT,
    is_public BOOLEAN DEFAULT 0

);