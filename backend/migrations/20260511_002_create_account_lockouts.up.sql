CREATE TABLE account_lockouts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    locked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    unlock_at TIMESTAMP NOT NULL,
    unlock_token_hash VARCHAR(255),
    unlock_token_expires_at TIMESTAMP,
    unlock_token_used_at TIMESTAMP,
    reason VARCHAR(255),
    lockout_count INTEGER NOT NULL DEFAULT 1,
    unlocked_at TIMESTAMP,
    unlocked_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_account_lockouts_user_id ON account_lockouts(user_id);
CREATE INDEX idx_account_lockouts_active ON account_lockouts(user_id, unlock_at) WHERE unlocked_at IS NULL;
CREATE UNIQUE INDEX idx_account_lockouts_unlock_token ON account_lockouts(unlock_token_hash) WHERE unlock_token_hash IS NOT NULL;
