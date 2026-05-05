-- +migrate Up

-- Admin approval tracking for pending user registrations
CREATE TABLE user_approvals (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
	reviewed_at TIMESTAMP,
	rejection_reason TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_approvals_user_id ON user_approvals(user_id);
CREATE INDEX idx_user_approvals_status ON user_approvals(status);
