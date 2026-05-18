-- +migrate Down

DELETE FROM role_permissions
WHERE resource_type IS NULL
  AND permission IN (
    'devices:read',
    'devices:write',
    'devices:delete',
    'devices:admin',
    'auth:admin:read',
    'auth:admin:write',
    'ipam:subnet_request:submit',
    'ipam:subnet_request:review',
    'ipam:customer:list',
    'ipam:customer:read',
    'ipam:customer:write',
    'ipam:customer:delete',
    'ipam:autonomous_system:list',
    'ipam:autonomous_system:read',
    'ipam:autonomous_system:write',
    'ipam:autonomous_system:delete'
  );
