-- +migrate Up

-- Backfill ipam:nameserver:list and ipam:nameserver:read into every role that
-- already has any IPAM permission (mirrors the location backfill pattern).
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:nameserver:list'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:nameserver:list'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:nameserver:read'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:nameserver:read'
        AND rp2.resource_type IS NULL
  );

-- Backfill write/delete to roles that already manage subnets.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:nameserver:write'
FROM role_permissions
WHERE permission = 'ipam:subnet:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:nameserver:write'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:nameserver:delete'
FROM role_permissions
WHERE permission = 'ipam:subnet:delete'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:nameserver:delete'
        AND rp2.resource_type IS NULL
  );
