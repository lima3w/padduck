-- +migrate Up
CREATE TABLE IF NOT EXISTS scan_profiles (
    id               BIGSERIAL PRIMARY KEY,
    name             TEXT NOT NULL,
    description      TEXT,
    scan_type        TEXT NOT NULL DEFAULT 'ping' CHECK (scan_type IN ('ping','snmp','ping+snmp')),
    ping_concurrency INT  NOT NULL DEFAULT 20 CHECK (ping_concurrency BETWEEN 1 AND 100),
    tcp_ports        TEXT,
    dns_lookup       BOOLEAN NOT NULL DEFAULT false,
    snmp_community   TEXT,
    snmp_version     TEXT NOT NULL DEFAULT 'v2c' CHECK (snmp_version IN ('v2c','v3')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE subnets ADD COLUMN IF NOT EXISTS scan_profile_id BIGINT REFERENCES scan_profiles(id) ON DELETE SET NULL;
