-- +migrate Up
CREATE TABLE IF NOT EXISTS secrets (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL DEFAULT 1 REFERENCES organizations(id),
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    current_version INT NOT NULL DEFAULT 1,
    created_by      BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    rotated_at      TIMESTAMPTZ,
    UNIQUE (organization_id, name)
);

-- One row per version. Each value is encrypted with its own data key (DEK);
-- the DEK is encrypted with the application key (KEK) and stored alongside.
CREATE TABLE IF NOT EXISTS secret_versions (
    id              BIGSERIAL PRIMARY KEY,
    secret_id       BIGINT NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    version         INT NOT NULL,
    encrypted_dek   BYTEA NOT NULL,
    encrypted_value BYTEA NOT NULL,
    created_by      BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (secret_id, version)
);
