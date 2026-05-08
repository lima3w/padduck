-- +migrate Up

-- Backfill ipam:location:list into every role that has any IPAM permission.
-- Required because location permissions were added after roles may have been created.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:location:list'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp2
    WHERE rp2.role_id = role_permissions.role_id
      AND rp2.permission = 'ipam:location:list'
      AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:location:read'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp2
    WHERE rp2.role_id = role_permissions.role_id
      AND rp2.permission = 'ipam:location:read'
      AND rp2.resource_type IS NULL
  );

-- Backfill write/delete to roles that already manage devices.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:location:write'
FROM role_permissions
WHERE permission = 'ipam:device:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp2
    WHERE rp2.role_id = role_permissions.role_id
      AND rp2.permission = 'ipam:location:write'
      AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:location:delete'
FROM role_permissions
WHERE permission = 'ipam:device:delete'
  AND resource_type IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp2
    WHERE rp2.role_id = role_permissions.role_id
      AND rp2.permission = 'ipam:location:delete'
      AND rp2.resource_type IS NULL
  );
