-- +migrate Up

-- #248: ping_concurrency on scan_jobs
ALTER TABLE scan_jobs ADD COLUMN IF NOT EXISTS ping_concurrency INT NOT NULL DEFAULT 20 CHECK (ping_concurrency BETWEEN 1 AND 100);

-- #248: notify_on_change on scan_jobs (also used by #211)
ALTER TABLE scan_jobs ADD COLUMN IF NOT EXISTS notify_on_change BOOLEAN NOT NULL DEFAULT false;

-- #210: scan_type on scan_jobs
ALTER TABLE scan_jobs ADD COLUMN IF NOT EXISTS scan_type TEXT NOT NULL DEFAULT 'ping' CHECK (scan_type IN ('ping','snmp','ping+snmp'));

-- #214: port_open JSONB on ip_addresses
ALTER TABLE ip_addresses ADD COLUMN IF NOT EXISTS port_open JSONB;

-- #211: scan_runs table
CREATE TABLE IF NOT EXISTS scan_runs (
    id            BIGSERIAL PRIMARY KEY,
    scan_job_id   BIGINT NOT NULL REFERENCES scan_jobs(id) ON DELETE CASCADE,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at   TIMESTAMPTZ,
    new_count     INT NOT NULL DEFAULT 0,
    gone_count    INT NOT NULL DEFAULT 0,
    changed_count INT NOT NULL DEFAULT 0
);

-- #211: scan_run_changes table
CREATE TABLE IF NOT EXISTS scan_run_changes (
    id          BIGSERIAL PRIMARY KEY,
    run_id      BIGINT NOT NULL REFERENCES scan_runs(id) ON DELETE CASCADE,
    ip_address  TEXT NOT NULL,
    change_type TEXT NOT NULL CHECK (change_type IN ('new','gone','changed')),
    scanned_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- #212: scan_agents table
CREATE TABLE IF NOT EXISTS scan_agents (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    last_seen  TIMESTAMPTZ,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- #212: agent_id on scan_jobs
ALTER TABLE scan_jobs ADD COLUMN IF NOT EXISTS agent_id BIGINT REFERENCES scan_agents(id) ON DELETE SET NULL;
