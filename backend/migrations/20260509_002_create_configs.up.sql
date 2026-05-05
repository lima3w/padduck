-- +migrate Up

-- Key-value store for application configuration
CREATE TABLE configs (
	key VARCHAR(100) PRIMARY KEY,
	value TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default registration and email configuration
INSERT INTO configs (key, value) VALUES
	('registration_enabled', 'true'),
	('require_email_verification', 'false'),
	('require_admin_approval', 'false'),
	('smtp_host', ''),
	('smtp_port', '587'),
	('smtp_username', ''),
	('smtp_password', ''),
	('smtp_from', ''),
	('smtp_tls', 'true');
