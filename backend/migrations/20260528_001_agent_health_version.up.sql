-- +migrate Up
ALTER TABLE scan_agents ADD COLUMN IF NOT EXISTS version TEXT;
ALTER TABLE scan_agents ADD COLUMN IF NOT EXISTS capabilities JSONB;
ALTER TABLE scan_agents ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'unknown' CHECK (status IN ('healthy','degraded','offline','unknown'));
ALTER TABLE scan_agents ADD COLUMN IF NOT EXISTS last_error TEXT;
