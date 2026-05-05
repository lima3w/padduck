-- +migrate Down

DROP INDEX IF EXISTS idx_users_state;
ALTER TABLE users DROP COLUMN IF EXISTS state;
