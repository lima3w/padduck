-- +migrate Down

DELETE FROM role_permissions WHERE permission IN (
    'ipam:vlan_group:list',
    'ipam:vlan_group:read',
    'ipam:vlan_group:write',
    'ipam:vlan_group:delete'
);
