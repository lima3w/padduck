-- +migrate Down
ALTER TABLE scan_agents DROP COLUMN IF EXISTS expires_at;
