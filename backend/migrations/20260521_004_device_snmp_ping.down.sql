-- +migrate Down
ALTER TABLE devices DROP COLUMN IF EXISTS is_online;
ALTER TABLE devices DROP COLUMN IF EXISTS last_ping_at;
ALTER TABLE devices DROP COLUMN IF EXISTS snmp_v3_priv_pass;
ALTER TABLE devices DROP COLUMN IF EXISTS snmp_v3_priv_proto;
ALTER TABLE devices DROP COLUMN IF EXISTS snmp_v3_auth_pass;
ALTER TABLE devices DROP COLUMN IF EXISTS snmp_v3_auth_proto;
ALTER TABLE devices DROP COLUMN IF EXISTS snmp_v3_user;
