-- +migrate Up

CREATE TABLE security_notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    ip_address INET,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_security_notifications_user_type_time ON security_notifications(user_id, notification_type, sent_at DESC);
