-- +migrate Up
CREATE TABLE IF NOT EXISTS device_fingerprints (
    id               BIGSERIAL PRIMARY KEY,
    device_id        BIGINT NOT NULL UNIQUE REFERENCES devices(id) ON DELETE CASCADE,
    open_ports       JSONB,
    os_guess         TEXT,
    vendor_guess     TEXT,
    confidence_score FLOAT NOT NULL DEFAULT 0 CHECK (confidence_score BETWEEN 0 AND 1),
    evidence         JSONB,
    last_updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
