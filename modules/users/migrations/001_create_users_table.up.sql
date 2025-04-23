-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    last_login TIMESTAMPTZ
);

-- Optional: Add index for faster email lookups if not created by UNIQUE constraint
-- CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Optional: Add index for soft delete
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at); 