-- +migrate Down
ALTER TABLE ip_addresses ADD COLUMN IF NOT EXISTS assigned_to VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_ip_addresses_assigned_to ON ip_addresses(assigned_to);
