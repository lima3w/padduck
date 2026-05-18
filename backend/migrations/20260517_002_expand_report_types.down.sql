-- +migrate Down

ALTER TABLE scheduled_reports
    DROP CONSTRAINT IF EXISTS scheduled_reports_report_type_check;

ALTER TABLE scheduled_reports
    ADD CONSTRAINT scheduled_reports_report_type_check
    CHECK (report_type IN ('utilisation_summary','inactive_ips','new_allocations','full_audit'));
