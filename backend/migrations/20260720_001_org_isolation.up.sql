-- +migrate Up
-- Organization isolation: scope tenant tables to an organization.
-- Existing rows belong to the default organization (id 1).

ALTER TABLE networks ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_sections_org ON networks (organization_id);
ALTER TABLE subnets ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_subnets_org ON subnets (organization_id);
ALTER TABLE ip_addresses ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_ip_addresses_org ON ip_addresses (organization_id);
ALTER TABLE vlans ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_vlans_org ON vlans (organization_id);
ALTER TABLE vlan_groups ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_vlan_groups_org ON vlan_groups (organization_id);
ALTER TABLE vlan_domains ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_vlan_domains_org ON vlan_domains (organization_id);
ALTER TABLE vrfs ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_vrfs_org ON vrfs (organization_id);
ALTER TABLE locations ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_locations_org ON locations (organization_id);
ALTER TABLE racks ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_racks_org ON racks (organization_id);
ALTER TABLE customers ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_customers_org ON customers (organization_id);
ALTER TABLE nameservers ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_nameservers_org ON nameservers (organization_id);
ALTER TABLE devices ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_devices_org ON devices (organization_id);
ALTER TABLE device_interfaces ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_device_interfaces_org ON device_interfaces (organization_id);
ALTER TABLE custom_field_definitions ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_custom_field_definitions_org ON custom_field_definitions (organization_id);
ALTER TABLE custom_field_values ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_custom_field_values_org ON custom_field_values (organization_id);
ALTER TABLE autonomous_systems ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_autonomous_systems_org ON autonomous_systems (organization_id);
ALTER TABLE nat_rules ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_nat_rules_org ON nat_rules (organization_id);
ALTER TABLE dhcp_servers ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_dhcp_servers_org ON dhcp_servers (organization_id);
ALTER TABLE dhcp_leases ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_dhcp_leases_org ON dhcp_leases (organization_id);
ALTER TABLE firewall_zones ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_firewall_zones_org ON firewall_zones (organization_id);
ALTER TABLE firewall_zone_mappings ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_firewall_zone_mappings_org ON firewall_zone_mappings (organization_id);
ALTER TABLE circuit_providers ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_circuit_providers_org ON circuit_providers (organization_id);
ALTER TABLE physical_circuits ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_physical_circuits_org ON physical_circuits (organization_id);
ALTER TABLE logical_circuits ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_logical_circuits_org ON logical_circuits (organization_id);
ALTER TABLE customer_associations ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_customer_associations_org ON customer_associations (organization_id);
ALTER TABLE scan_profiles ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_scan_profiles_org ON scan_profiles (organization_id);
ALTER TABLE scan_jobs ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_scan_jobs_org ON scan_jobs (organization_id);
ALTER TABLE scan_results ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_scan_results_org ON scan_results (organization_id);
ALTER TABLE scan_runs ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_scan_runs_org ON scan_runs (organization_id);
ALTER TABLE scan_run_changes ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_scan_run_changes_org ON scan_run_changes (organization_id);
ALTER TABLE scan_agents ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_scan_agents_org ON scan_agents (organization_id);
ALTER TABLE ip_requests ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_ip_requests_org ON ip_requests (organization_id);
ALTER TABLE subnet_requests ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_subnet_requests_org ON subnet_requests (organization_id);
ALTER TABLE request_comments ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_request_comments_org ON request_comments (organization_id);
ALTER TABLE discovery_conflicts ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_discovery_conflicts_org ON discovery_conflicts (organization_id);
ALTER TABLE automation_policies ADD COLUMN organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id);
CREATE INDEX idx_automation_policies_org ON automation_policies (organization_id);

ALTER TABLE ip_tags ADD COLUMN organization_id BIGINT REFERENCES organizations(id);
CREATE INDEX idx_ip_tags_org ON ip_tags (organization_id);
ALTER TABLE device_types ADD COLUMN organization_id BIGINT REFERENCES organizations(id);
CREATE INDEX idx_device_types_org ON device_types (organization_id);

-- Names that were globally unique become unique per organization.
ALTER TABLE vrfs DROP CONSTRAINT IF EXISTS vrfs_name_key;
ALTER TABLE vrfs ADD CONSTRAINT vrfs_org_name_key UNIQUE (organization_id, name);
ALTER TABLE vlan_domains DROP CONSTRAINT IF EXISTS vlan_domains_name_key;
ALTER TABLE vlan_domains ADD CONSTRAINT vlan_domains_org_name_key UNIQUE (organization_id, name);
ALTER TABLE vlan_groups DROP CONSTRAINT IF EXISTS vlan_groups_name_key;
ALTER TABLE vlan_groups ADD CONSTRAINT vlan_groups_org_name_key UNIQUE (organization_id, name);
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_name_key;
ALTER TABLE customers ADD CONSTRAINT customers_org_name_key UNIQUE (organization_id, name);
ALTER TABLE autonomous_systems DROP CONSTRAINT IF EXISTS autonomous_systems_asn_key;
ALTER TABLE autonomous_systems ADD CONSTRAINT autonomous_systems_org_asn_key UNIQUE (organization_id, asn);
ALTER TABLE nat_rules DROP CONSTRAINT IF EXISTS nat_rules_name_key;
ALTER TABLE nat_rules ADD CONSTRAINT nat_rules_org_name_key UNIQUE (organization_id, name);
ALTER TABLE dhcp_servers DROP CONSTRAINT IF EXISTS dhcp_servers_name_key;
ALTER TABLE dhcp_servers ADD CONSTRAINT dhcp_servers_org_name_key UNIQUE (organization_id, name);
ALTER TABLE circuit_providers DROP CONSTRAINT IF EXISTS circuit_providers_name_key;
ALTER TABLE circuit_providers ADD CONSTRAINT circuit_providers_org_name_key UNIQUE (organization_id, name);
ALTER TABLE logical_circuits DROP CONSTRAINT IF EXISTS logical_circuits_name_key;
ALTER TABLE logical_circuits ADD CONSTRAINT logical_circuits_org_name_key UNIQUE (organization_id, name);
ALTER TABLE firewall_zones DROP CONSTRAINT IF EXISTS firewall_zones_name_key;
ALTER TABLE firewall_zones ADD CONSTRAINT firewall_zones_org_name_key UNIQUE (organization_id, name);
ALTER TABLE vlans DROP CONSTRAINT IF EXISTS vlans_vlan_id_key;
ALTER TABLE vlans ADD CONSTRAINT vlans_org_vlan_id_key UNIQUE (organization_id, vlan_id);
ALTER TABLE ip_tags DROP CONSTRAINT IF EXISTS ip_tags_name_key;
CREATE UNIQUE INDEX ip_tags_org_name_key ON ip_tags (COALESCE(organization_id, 0), name);
ALTER TABLE device_types DROP CONSTRAINT IF EXISTS device_types_name_key;
CREATE UNIQUE INDEX device_types_org_name_key ON device_types (COALESCE(organization_id, 0), name);

-- Graph rows created by the sync carried no organization; they belong to the default org.
UPDATE topology_nodes SET organization_id = 1 WHERE organization_id IS NULL;
UPDATE topology_edges SET organization_id = 1 WHERE organization_id IS NULL;

-- Users created after the organizations migration carried no organization. Every
-- user belongs to exactly one organization; new users default to the default org.
UPDATE users SET organization_id = 1 WHERE organization_id IS NULL;
ALTER TABLE users ALTER COLUMN organization_id SET DEFAULT 1;
ALTER TABLE users ALTER COLUMN organization_id SET NOT NULL;
