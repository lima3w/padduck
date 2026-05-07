-- +migrate Up
ALTER TABLE api_tokens
    ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL DEFAULT 'write',
    ADD COLUMN IF NOT EXISTS usage_count BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_used_ip TEXT,
    ADD COLUMN IF NOT EXISTS rotation_grace_expires_at TIMESTAMP;
