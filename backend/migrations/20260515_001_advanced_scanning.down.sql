-- +migrate Down

ALTER TABLE scan_jobs DROP COLUMN IF EXISTS agent_id;
DROP TABLE IF EXISTS scan_agents;
DROP TABLE IF EXISTS scan_run_changes;
DROP TABLE IF EXISTS scan_runs;
ALTER TABLE ip_addresses DROP COLUMN IF EXISTS port_open;
ALTER TABLE scan_jobs DROP COLUMN IF EXISTS scan_type;
ALTER TABLE scan_jobs DROP COLUMN IF EXISTS notify_on_change;
ALTER TABLE scan_jobs DROP COLUMN IF EXISTS ping_concurrency;
