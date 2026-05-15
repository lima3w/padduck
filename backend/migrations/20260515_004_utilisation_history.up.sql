-- +migrate Up
CREATE TABLE subnet_utilisation_history (
    id BIGSERIAL PRIMARY KEY,
    subnet_id BIGINT NOT NULL REFERENCES subnets(id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    used_count INTEGER NOT NULL,
    total_count INTEGER NOT NULL,
    utilisation_pct NUMERIC(5,2) NOT NULL
);
CREATE INDEX idx_util_history_subnet_time ON subnet_utilisation_history(subnet_id, recorded_at DESC);
