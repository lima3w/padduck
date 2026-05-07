-- +migrate Down
ALTER TABLE api_tokens
    DROP COLUMN IF EXISTS scope,
    DROP COLUMN IF EXISTS usage_count,
    DROP COLUMN IF EXISTS last_used_ip,
    DROP COLUMN IF EXISTS rotation_grace_expires_at;
