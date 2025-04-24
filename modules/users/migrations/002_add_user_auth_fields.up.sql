-- Add user_type and is_active columns to users table
ALTER TABLE users ADD COLUMN user_type VARCHAR(50) NOT NULL DEFAULT 'ADMIN';
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT TRUE; 