-- +migrate Down
UPDATE customer_associations SET object_type = 'section' WHERE object_type = 'network';
ALTER TABLE customer_associations DROP CONSTRAINT customer_associations_object_check;
ALTER TABLE customer_associations ADD CONSTRAINT customer_associations_object_check
    CHECK (object_type IN ('section', 'subnet', 'ip_address', 'device', 'rack', 'location', 'vlan', 'vrf', 'nat_rule', 'dhcp_server', 'dhcp_lease', 'physical_circuit', 'logical_circuit'));
