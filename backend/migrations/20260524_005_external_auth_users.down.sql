-- +migrate Down

DROP INDEX IF EXISTS idx_users_external_auth;

ALTER TABLE users
    DROP COLUMN IF EXISTS external_auth_provider,
    DROP COLUMN IF EXISTS external_auth_id;