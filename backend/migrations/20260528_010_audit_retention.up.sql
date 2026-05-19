CREATE TABLE IF NOT EXISTS audit_retention_settings (
    id BIGSERIAL PRIMARY KEY,
    retention_days INT NOT NULL DEFAULT 365 CHECK (retention_days >= 30),
    archive_enabled BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO audit_retention_settings (retention_days, archive_enabled)
VALUES (365, false) ON CONFLICT DO NOTHING;
