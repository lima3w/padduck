-- +migrate Up

-- Add state column to users table for registration workflow
ALTER TABLE users ADD COLUMN state VARCHAR(50) NOT NULL DEFAULT 'active';

-- Create index for state lookups
CREATE INDEX idx_users_state ON users(state);
