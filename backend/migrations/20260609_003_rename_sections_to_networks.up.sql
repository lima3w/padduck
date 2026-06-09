-- +migrate Up
ALTER TABLE sections RENAME TO networks;
ALTER TABLE subnets RENAME COLUMN section_id TO network_id;
ALTER INDEX IF EXISTS idx_sections_created_by RENAME TO idx_networks_created_by;
ALTER INDEX IF EXISTS idx_subnets_section_id RENAME TO idx_subnets_network_id;
