-- +migrate Up
CREATE TABLE IF NOT EXISTS scan_retention_settings (
    id                BIGSERIAL PRIMARY KEY,
    raw_history_days  INT NOT NULL DEFAULT 90 CHECK (raw_history_days >= 1),
    rollup_enabled    BOOLEAN NOT NULL DEFAULT false,
    rollup_after_days INT NOT NULL DEFAULT 30 CHECK (rollup_after_days >= 1),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Insert default row
INSERT INTO scan_retention_settings (raw_history_days, rollup_enabled, rollup_after_days)
VALUES (90, false, 30)
ON CONFLICT DO NOTHING;
