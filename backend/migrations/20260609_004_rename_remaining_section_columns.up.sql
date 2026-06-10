-- +migrate Up
ALTER TABLE devices RENAME COLUMN section_id TO network_id;
ALTER TABLE subnet_requests RENAME COLUMN section_id TO network_id;
ALTER INDEX IF EXISTS idx_subnet_requests_section RENAME TO idx_subnet_requests_network;
DROP INDEX IF EXISTS idx_subnets_section_network;
CREATE INDEX IF NOT EXISTS idx_subnets_network_cidr ON subnets(network_id, network_address);
