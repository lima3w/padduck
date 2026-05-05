-- +migrate Up

-- Add role column to users table
ALTER TABLE users ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'user';

-- Add index on role for efficient querying
CREATE INDEX idx_users_role ON users(role);

-- Create role types (admin, user, viewer)
-- Role permissions:
-- admin: full access to all resources
-- user: read/write on sections, subnets, IP addresses
-- viewer: read-only access to sections, subnets, IP addresses
