-- +migrate Down

DELETE FROM role_permissions WHERE permission IN (
  'ipam:firewall:list', 'ipam:firewall:read', 'ipam:firewall:write', 'ipam:firewall:delete'
);

DELETE FROM configs WHERE key = 'feature_firewall_enabled';

DROP TABLE IF EXISTS firewall_zone_mappings;
DROP TABLE IF EXISTS firewall_zones;
