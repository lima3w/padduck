-- +migrate Down
DELETE FROM roles WHERE is_system = TRUE;
