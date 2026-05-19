-- +migrate Up

CREATE INDEX IF NOT EXISTS idx_ip_addresses_subnet_status_address
    ON ip_addresses(subnet_id, status, address);

CREATE INDEX IF NOT EXISTS idx_ip_addresses_status_last_seen
    ON ip_addresses(status, last_seen)
    WHERE status = 'assigned';

CREATE INDEX IF NOT EXISTS idx_ip_addresses_device_id
    ON ip_addresses(device_id)
    WHERE device_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ip_addresses_dns_name
    ON ip_addresses(dns_name)
    WHERE dns_name IS NOT NULL AND dns_name <> '';

CREATE INDEX IF NOT EXISTS idx_ip_addresses_address_status
    ON ip_addresses(address, status);

CREATE INDEX IF NOT EXISTS idx_subnets_section_network
    ON subnets(section_id, network_address);

CREATE INDEX IF NOT EXISTS idx_subnets_location_network
    ON subnets(location_id, network_address)
    WHERE location_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_subnets_alert_threshold
    ON subnets(network_address)
    WHERE alert_threshold_pct IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_dashboard_activity
    ON audit_logs(created_at DESC)
    WHERE action IN ('ip_assigned', 'ip_released', 'subnet_created', 'subnet_deleted', 'subnet_updated');

CREATE INDEX IF NOT EXISTS idx_devices_hostname
    ON devices(hostname)
    WHERE hostname IS NOT NULL AND hostname <> '';

CREATE INDEX IF NOT EXISTS idx_devices_last_ping_at
    ON devices(last_ping_at);

CREATE INDEX IF NOT EXISTS idx_scan_jobs_active_last_run
    ON scan_jobs(is_active, last_run_at)
    WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_scan_runs_job_finished
    ON scan_runs(scan_job_id, finished_at DESC)
    WHERE finished_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_subnet_requests_status_created
    ON subnet_requests(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ip_requests_status_created
    ON ip_requests(status, created_at DESC);
