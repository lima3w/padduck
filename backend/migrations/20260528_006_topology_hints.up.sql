-- +migrate Up
CREATE TABLE IF NOT EXISTS topology_hints (
    id BIGSERIAL PRIMARY KEY,
    source_type TEXT NOT NULL,
    source_id BIGINT NOT NULL,
    target_type TEXT NOT NULL,
    target_id BIGINT NOT NULL,
    hint_type TEXT NOT NULL,
    confidence_score FLOAT NOT NULL DEFAULT 0 CHECK (confidence_score BETWEEN 0 AND 1),
    evidence TEXT,
    status TEXT NOT NULL DEFAULT 'suggested' CHECK (status IN ('suggested','confirmed','dismissed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_topology_hints_source ON topology_hints(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_topology_hints_status ON topology_hints(status);
ALTER TABLE topology_hints ADD CONSTRAINT uq_topology_hints UNIQUE (source_type, source_id, target_type, target_id, hint_type);
