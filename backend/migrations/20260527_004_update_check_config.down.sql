-- +migrate Down

DELETE FROM configs
WHERE key IN (
	'update_check_enabled',
	'update_check_url',
	'update_check_token'
);
