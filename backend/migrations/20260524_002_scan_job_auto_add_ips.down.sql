-- +migrate Down
ALTER TABLE scan_jobs DROP COLUMN IF EXISTS auto_add_ips;
