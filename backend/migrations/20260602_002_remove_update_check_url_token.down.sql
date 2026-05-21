-- +migrate Down

INSERT INTO configs (key, value) VALUES
	('update_check_url', ''),
	('update_check_token', '')
ON CONFLICT (key) DO NOTHING;
