-- +migrate Down
ALTER TABLE ip_tags RENAME COLUMN color TO colour;
ALTER TABLE vlan_groups RENAME COLUMN color TO colour;

ALTER TABLE subnet_utilization_history RENAME COLUMN utilization_pct TO utilisation_pct;
ALTER TABLE subnet_utilization_history RENAME TO subnet_utilisation_history;
ALTER SEQUENCE IF EXISTS subnet_utilization_history_id_seq RENAME TO subnet_utilisation_history_id_seq;
ALTER INDEX IF EXISTS subnet_utilization_history_pkey RENAME TO subnet_utilisation_history_pkey;

ALTER TABLE scheduled_reports DROP CONSTRAINT IF EXISTS scheduled_reports_report_type_check;
UPDATE scheduled_reports SET report_type = 'utilisation_summary' WHERE report_type = 'utilization_summary';
ALTER TABLE scheduled_reports
    ADD CONSTRAINT scheduled_reports_report_type_check
    CHECK (report_type IN (
        'utilisation_summary',
        'inactive_ips',
        'new_allocations',
        'full_audit',
        'subnet_gaps',
        'vlan_assignment',
        'ip_age',
        'dns_audit'
    ));

UPDATE configs SET key = 'utilisation_snapshot_interval_hours'
 WHERE key = 'utilization_snapshot_interval_hours'
   AND NOT EXISTS (SELECT 1 FROM configs WHERE key = 'utilisation_snapshot_interval_hours');
DELETE FROM configs WHERE key = 'utilization_snapshot_interval_hours';
