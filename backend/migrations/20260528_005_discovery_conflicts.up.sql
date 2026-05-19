CREATE TABLE IF NOT EXISTS discovery_conflicts (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    field_name TEXT NOT NULL,
    current_value TEXT,
    discovered_value TEXT NOT NULL,
    confidence_score FLOAT NOT NULL DEFAULT 0 CHECK (confidence_score BETWEEN 0 AND 1),
    source TEXT NOT NULL DEFAULT 'scan',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','rejected')),
    reviewed_by TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_discovery_conflicts_device ON discovery_conflicts(device_id);
CREATE INDEX IF NOT EXISTS idx_discovery_conflicts_status ON discovery_conflicts(status);
