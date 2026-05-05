-- +migrate Up

-- Email verification tokens for new user registrations
CREATE TABLE email_verifications (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token_hash VARCHAR(255) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	used_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX idx_email_verifications_token_hash ON email_verifications(token_hash);
