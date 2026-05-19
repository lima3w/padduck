-- +migrate Down
ALTER TABLE scan_agents DROP COLUMN IF EXISTS version;
ALTER TABLE scan_agents DROP COLUMN IF EXISTS capabilities;
ALTER TABLE scan_agents DROP COLUMN IF EXISTS status;
ALTER TABLE scan_agents DROP COLUMN IF EXISTS last_error;
