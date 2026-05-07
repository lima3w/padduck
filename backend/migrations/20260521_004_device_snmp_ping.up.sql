-- +migrate Up
ALTER TABLE devices ADD COLUMN snmp_v3_user VARCHAR(100);
ALTER TABLE devices ADD COLUMN snmp_v3_auth_proto VARCHAR(20);
ALTER TABLE devices ADD COLUMN snmp_v3_auth_pass VARCHAR(500);
ALTER TABLE devices ADD COLUMN snmp_v3_priv_proto VARCHAR(20);
ALTER TABLE devices ADD COLUMN snmp_v3_priv_pass VARCHAR(500);
ALTER TABLE devices ADD COLUMN last_ping_at TIMESTAMPTZ;
ALTER TABLE devices ADD COLUMN is_online BOOLEAN NOT NULL DEFAULT false;
