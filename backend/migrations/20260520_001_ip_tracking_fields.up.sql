ALTER TABLE ip_addresses ADD COLUMN last_seen TIMESTAMPTZ;
ALTER TABLE ip_addresses ADD COLUMN mac_address VARCHAR(17);
ALTER TABLE ip_addresses ADD COLUMN ptr_record VARCHAR(253);
