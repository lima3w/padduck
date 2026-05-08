-- +migrate Down

DELETE FROM role_permissions
WHERE permission IN (
    'ipam:location:list',
    'ipam:location:read',
    'ipam:location:write',
    'ipam:location:delete'
);
