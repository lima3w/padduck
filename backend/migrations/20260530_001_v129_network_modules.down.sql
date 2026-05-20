-- +migrate Down

DELETE FROM role_permissions WHERE permission IN (
  'ipam:nat:list', 'ipam:nat:read', 'ipam:nat:write', 'ipam:nat:delete',
  'ipam:dhcp:list', 'ipam:dhcp:read', 'ipam:dhcp:write', 'ipam:dhcp:delete',
  'ipam:circuit:list', 'ipam:circuit:read', 'ipam:circuit:write', 'ipam:circuit:delete'
);

DELETE FROM configs WHERE key IN ('feature_nat_enabled', 'feature_dhcp_enabled', 'feature_circuits_enabled');

DROP TABLE IF EXISTS customer_associations;
DROP TABLE IF EXISTS logical_circuits;
DROP TABLE IF EXISTS physical_circuits;
DROP TABLE IF EXISTS circuit_providers;
DROP TABLE IF EXISTS dhcp_leases;
DROP TABLE IF EXISTS dhcp_servers;
DROP TABLE IF EXISTS nat_rules;

ALTER TABLE locations
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS contact_phone,
  DROP COLUMN IF EXISTS contact_email,
  DROP COLUMN IF EXISTS contact_name,
  DROP COLUMN IF EXISTS time_zone,
  DROP COLUMN IF EXISTS facility_code,
  DROP COLUMN IF EXISTS country,
  DROP COLUMN IF EXISTS region,
  DROP COLUMN IF EXISTS city;
