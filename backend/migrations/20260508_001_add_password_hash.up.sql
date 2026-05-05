-- +migrate Up

-- Add password_hash column to users table
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);

-- Add last_login_at to track user activity
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP;
