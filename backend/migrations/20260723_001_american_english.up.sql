-- +migrate Up
-- v2.0.0: normalize British spellings to American English (#146).

ALTER TABLE ip_tags RENAME COLUMN colour TO color;
ALTER TABLE vlan_groups RENAME COLUMN colour TO color;

ALTER TABLE subnet_utilisation_history RENAME TO subnet_utilization_history;
ALTER TABLE subnet_utilization_history RENAME COLUMN utilisation_pct TO utilization_pct;
ALTER SEQUENCE IF EXISTS subnet_utilisation_history_id_seq RENAME TO subnet_utilization_history_id_seq;
ALTER INDEX IF EXISTS subnet_utilisation_history_pkey RENAME TO subnet_utilization_history_pkey;

-- The report type is stored and exposed through the scheduled reports API.
ALTER TABLE scheduled_reports DROP CONSTRAINT IF EXISTS scheduled_reports_report_type_check;
UPDATE scheduled_reports SET report_type = 'utilization_summary' WHERE report_type = 'utilisation_summary';
ALTER TABLE scheduled_reports
    ADD CONSTRAINT scheduled_reports_report_type_check
    CHECK (report_type IN (
        'utilization_summary',
        'inactive_ips',
        'new_allocations',
        'full_audit',
        'subnet_gaps',
        'vlan_assignment',
        'ip_age',
        'dns_audit'
    ));

-- Admin config key. Keep any value an operator already set.
UPDATE configs SET key = 'utilization_snapshot_interval_hours'
 WHERE key = 'utilisation_snapshot_interval_hours'
   AND NOT EXISTS (SELECT 1 FROM configs WHERE key = 'utilization_snapshot_interval_hours');
DELETE FROM configs WHERE key = 'utilisation_snapshot_interval_hours';
