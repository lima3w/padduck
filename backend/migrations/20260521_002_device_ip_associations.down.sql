-- +migrate Down
ALTER TABLE ip_addresses DROP COLUMN IF EXISTS is_primary;
ALTER TABLE ip_addresses DROP COLUMN IF EXISTS interface_name;
ALTER TABLE ip_addresses DROP COLUMN IF EXISTS device_id;
