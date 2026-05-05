-- +migrate Down

DELETE FROM configs WHERE key = 'app_url';
