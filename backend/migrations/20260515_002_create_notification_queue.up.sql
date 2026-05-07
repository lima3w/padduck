-- +migrate Up
CREATE TABLE notification_queue (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    template TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMP,
    sent_at TIMESTAMP,
    error_msg TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_queue_pending ON notification_queue(status, next_retry_at) WHERE status IN ('pending', 'retrying');
CREATE INDEX idx_notification_queue_user ON notification_queue(user_id, created_at);
