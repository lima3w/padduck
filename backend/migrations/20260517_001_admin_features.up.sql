-- +migrate Up

-- User suspension tracking
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS suspended_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS suspension_reason TEXT;

-- Privacy policy consent tracking
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS privacy_accepted_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS privacy_accepted_version TEXT;

-- GDPR deletion tracking
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS deletion_requested_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS anonymized_at TIMESTAMP;

-- Impersonation tracking on sessions
ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS is_impersonation BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS impersonated_by BIGINT REFERENCES users(id) ON DELETE SET NULL;

-- Seed privacy policy version config
INSERT INTO configs (key, value) VALUES ('privacy_policy_version', '1.0') ON CONFLICT (key) DO NOTHING;
