-- +migrate Up

-- Per-user MFA status
CREATE TABLE user_mfa_settings (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	totp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
	backup_codes_generated_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Encrypted TOTP secrets (AES-256-GCM)
CREATE TABLE user_totp_secrets (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	encrypted_secret BYTEA NOT NULL,
	verified BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Hashed one-time backup codes
CREATE TABLE user_backup_codes (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	code_hash VARCHAR(255) NOT NULL,
	used BOOLEAN NOT NULL DEFAULT FALSE,
	used_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_backup_codes_user_id ON user_backup_codes(user_id);

-- Short-lived challenges issued after password auth when MFA is enabled
CREATE TABLE mfa_challenges (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	challenge_hash VARCHAR(255) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	completed_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mfa_challenges_challenge_hash ON mfa_challenges(challenge_hash);
