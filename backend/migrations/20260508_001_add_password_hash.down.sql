-- +migrate Down

-- Remove password_hash and last_login_at columns
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
