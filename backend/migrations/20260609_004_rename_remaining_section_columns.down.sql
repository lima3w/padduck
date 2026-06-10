-- +migrate Down
DROP INDEX IF EXISTS idx_subnets_network_cidr;
CREATE INDEX IF NOT EXISTS idx_subnets_section_network ON subnets(network_id, network_address);
ALTER INDEX IF EXISTS idx_subnet_requests_network RENAME TO idx_subnet_requests_section;
ALTER TABLE subnet_requests RENAME COLUMN network_id TO section_id;
ALTER TABLE devices RENAME COLUMN network_id TO section_id;
