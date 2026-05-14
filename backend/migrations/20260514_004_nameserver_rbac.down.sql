-- +migrate Down

DELETE FROM role_permissions
WHERE permission IN (
    'ipam:nameserver:list',
    'ipam:nameserver:read',
    'ipam:nameserver:write',
    'ipam:nameserver:delete'
)
AND resource_type IS NULL;
