-- +migrate Down

DROP INDEX IF EXISTS idx_ip_requests_status_created;
DROP INDEX IF EXISTS idx_subnet_requests_status_created;
DROP INDEX IF EXISTS idx_scan_runs_job_finished;
DROP INDEX IF EXISTS idx_scan_jobs_active_last_run;
DROP INDEX IF EXISTS idx_devices_last_ping_at;
DROP INDEX IF EXISTS idx_devices_hostname;
DROP INDEX IF EXISTS idx_audit_logs_dashboard_activity;
DROP INDEX IF EXISTS idx_subnets_alert_threshold;
DROP INDEX IF EXISTS idx_subnets_location_network;
DROP INDEX IF EXISTS idx_subnets_section_network;
DROP INDEX IF EXISTS idx_ip_addresses_address_status;
DROP INDEX IF EXISTS idx_ip_addresses_dns_name;
DROP INDEX IF EXISTS idx_ip_addresses_device_id;
DROP INDEX IF EXISTS idx_ip_addresses_status_last_seen;
DROP INDEX IF EXISTS idx_ip_addresses_subnet_status_address;
