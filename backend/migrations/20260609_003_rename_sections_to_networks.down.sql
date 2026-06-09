ALTER TABLE networks RENAME TO sections;
ALTER TABLE subnets RENAME COLUMN network_id TO section_id;
ALTER INDEX IF EXISTS idx_networks_created_by RENAME TO idx_sections_created_by;
ALTER INDEX IF EXISTS idx_subnets_network_id RENAME TO idx_subnets_section_id;
