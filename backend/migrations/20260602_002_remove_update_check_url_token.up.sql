-- +migrate Up

DELETE FROM configs
WHERE key IN ('update_check_url', 'update_check_token');
