-- +migrate Up
ALTER TABLE ip_addresses ADD COLUMN device_id INT REFERENCES devices(id) ON DELETE SET NULL;
ALTER TABLE ip_addresses ADD COLUMN interface_name VARCHAR(100);
ALTER TABLE ip_addresses ADD COLUMN is_primary BOOLEAN NOT NULL DEFAULT false;
