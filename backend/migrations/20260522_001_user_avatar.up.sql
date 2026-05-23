-- +migrate Up

-- Add avatar support to users table.
-- avatar_source: 'gravatar' (default) or 'custom'
-- avatar_data: base64 data URL for custom avatars; NULL when using Gravatar
ALTER TABLE users
    ADD COLUMN avatar_source VARCHAR(20) NOT NULL DEFAULT 'gravatar',
    ADD COLUMN avatar_data   TEXT;
