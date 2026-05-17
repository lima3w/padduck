-- +migrate Up
CREATE TABLE autonomous_systems (
    id          BIGSERIAL PRIMARY KEY,
    asn         BIGINT NOT NULL UNIQUE,
    name        TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    type        TEXT NOT NULL DEFAULT 'external' CHECK (type IN ('internal', 'external')),
    rir         TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
