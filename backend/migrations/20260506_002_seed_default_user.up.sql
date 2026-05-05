-- +migrate Up

-- Insert default admin user if not exists
INSERT INTO users (username, email, role, created_at, updated_at)
VALUES ('admin', 'admin@localhost', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (username) DO NOTHING;
