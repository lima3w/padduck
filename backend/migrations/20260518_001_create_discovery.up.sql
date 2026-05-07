-- +migrate Up

CREATE TABLE scan_jobs (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    subnet_ids BIGINT[] NOT NULL DEFAULT '{}',
    schedule_cron TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scan_results (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL REFERENCES scan_jobs(id) ON DELETE CASCADE,
    subnet_id BIGINT NOT NULL REFERENCES subnets(id) ON DELETE CASCADE,
    ip_address_id BIGINT REFERENCES ip_addresses(id) ON DELETE SET NULL,
    ip_address TEXT NOT NULL,
    is_alive BOOLEAN NOT NULL DEFAULT FALSE,
    response_time_ms BIGINT,
    scanned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scan_results_job_id ON scan_results(job_id);
CREATE INDEX idx_scan_results_subnet_id ON scan_results(subnet_id);
CREATE INDEX idx_scan_results_scanned_at ON scan_results(scanned_at);
