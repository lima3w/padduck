-- +migrate Up

CREATE TABLE IF NOT EXISTS oauth2_configs (
    id                  BIGSERIAL PRIMARY KEY,
    enabled             BOOLEAN NOT NULL DEFAULT FALSE,
    provider_name       TEXT NOT NULL DEFAULT '',
    client_id           TEXT NOT NULL DEFAULT '',
    client_secret_enc   BYTEA NOT NULL DEFAULT '',
    discovery_url       TEXT NOT NULL DEFAULT '',
    authorization_url   TEXT NOT NULL DEFAULT '',
    token_url           TEXT NOT NULL DEFAULT '',
    userinfo_url        TEXT NOT NULL DEFAULT '',
    scopes              TEXT NOT NULL DEFAULT 'openid email profile',
    redirect_uri        TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oauth2_states (
    id           BIGSERIAL PRIMARY KEY,
    state        TEXT NOT NULL UNIQUE,
    redirect_uri TEXT NOT NULL DEFAULT '',
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);