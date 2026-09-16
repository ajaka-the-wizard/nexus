CREATE TYPE role_enum AS ENUM ('user','admin','superadmin','support');

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    full_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    verified BOOLEAN DEFAULT FALSE,
    paid BOOLEAN DEFAULT FALSE,
    active BOOLEAN DEFAULT TRUE,
    role role_enum NOT NULL DEFAULT 'user',
    password_hash TEXT,
    password_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_updated_at ON users(updated_at);