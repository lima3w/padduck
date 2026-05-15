-- +migrate Down

DELETE FROM role_permissions WHERE permission IN (
    'ipam:vlan_domain:list',
    'ipam:vlan_domain:read',
    'ipam:vlan_domain:write',
    'ipam:vlan_domain:delete'
);
