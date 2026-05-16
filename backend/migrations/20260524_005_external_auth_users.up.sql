-- +migrate Up

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS external_auth_provider TEXT,
    ADD COLUMN IF NOT EXISTS external_auth_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_external_auth
    ON users (external_auth_provider, external_auth_id)
    WHERE external_auth_provider IS NOT NULL AND external_auth_id IS NOT NULL;