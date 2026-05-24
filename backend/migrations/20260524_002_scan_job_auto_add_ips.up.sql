-- +migrate Up
ALTER TABLE scan_jobs ADD COLUMN IF NOT EXISTS auto_add_ips BOOLEAN NOT NULL DEFAULT true;
