-- +migrate Down

ALTER TABLE ip_addresses
DROP COLUMN assigned_at,
DROP COLUMN expires_at;
