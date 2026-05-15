-- +migrate Up
CREATE TABLE scheduled_reports (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN ('utilisation_summary','inactive_ips','new_allocations','full_audit')),
    schedule_cron VARCHAR(100) NOT NULL,
    recipient_emails JSONB NOT NULL DEFAULT '[]',
    filters JSONB NOT NULL DEFAULT '{}',
    format VARCHAR(10) NOT NULL DEFAULT 'csv' CHECK (format IN ('csv','pdf')),
    last_run_at TIMESTAMPTZ,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
