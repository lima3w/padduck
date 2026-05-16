-- +migrate Down

DELETE FROM role_permissions
WHERE permission IN (
    'ipam:location:write',
    'ipam:location:delete'
);
