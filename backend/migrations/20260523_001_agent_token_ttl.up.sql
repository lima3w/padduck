-- +migrate Up
ALTER TABLE scan_agents ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
