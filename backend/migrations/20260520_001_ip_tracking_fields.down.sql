-- +migrate Down

ALTER TABLE ip_addresses DROP COLUMN IF EXISTS ptr_record;
ALTER TABLE ip_addresses DROP COLUMN IF EXISTS mac_address;
ALTER TABLE ip_addresses DROP COLUMN IF EXISTS last_seen;
