-- +migrate Up

CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_name VARCHAR(255) NOT NULL DEFAULT 'Unknown Device',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    absolute_expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_absolute_expires_at ON sessions(absolute_expires_at);

-- Default session timeout config values
INSERT INTO config (key, value) VALUES
    ('session_idle_timeout_minutes', '60'),
    ('session_absolute_timeout_hours', '168')
ON CONFLICT (key) DO NOTHING;
