-- +migrate Up
-- Fresh installs start with the modules a small homelab rarely needs turned
-- off; admins can enable them under Admin > Settings. Existing installs are
-- left exactly as they are: the migration only acts while no user exists,
-- which is true only when the whole schema is being created from scratch.
UPDATE configs
SET value = 'false'
WHERE key IN (
    'feature_bgp_enabled',
    'feature_circuits_enabled',
    'feature_customers_enabled',
    'feature_firewall_enabled',
    'feature_nat_enabled',
    'feature_vrfs_enabled',
    'feature_racks_enabled'
)
AND NOT EXISTS (SELECT 1 FROM users);
