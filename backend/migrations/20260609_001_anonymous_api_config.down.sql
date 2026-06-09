-- +migrate Down

DELETE FROM configs WHERE key = 'anonymous_api_enabled';
