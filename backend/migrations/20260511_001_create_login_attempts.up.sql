-- +migrate Up

CREATE TABLE login_attempts (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    failure_reason VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_login_attempts_username_created ON login_attempts(username, created_at DESC);
CREATE INDEX idx_login_attempts_ip_created ON login_attempts(ip_address, created_at DESC);
CREATE INDEX idx_login_attempts_failures ON login_attempts(username, created_at DESC) WHERE success = false;
