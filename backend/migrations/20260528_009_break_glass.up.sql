CREATE TABLE IF NOT EXISTS break_glass_sessions (
    id BIGSERIAL PRIMARY KEY,
    initiated_by_user_id BIGINT NOT NULL REFERENCES users(id),
    justification TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    ended_by_user_id BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_break_glass_active ON break_glass_sessions(expires_at, ended_at);
