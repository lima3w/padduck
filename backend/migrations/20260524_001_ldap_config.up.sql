-- +migrate Up

CREATE TABLE IF NOT EXISTS ldap_configs (
    id                BIGSERIAL PRIMARY KEY,
    enabled           BOOLEAN NOT NULL DEFAULT FALSE,
    host              TEXT NOT NULL DEFAULT '',
    port              INTEGER NOT NULL DEFAULT 389,
    bind_dn           TEXT NOT NULL DEFAULT '',
    bind_password_enc BYTEA NOT NULL DEFAULT '',
    base_dn           TEXT NOT NULL DEFAULT '',
    user_filter       TEXT NOT NULL DEFAULT '(sAMAccountName=%s)',
    username_attr     TEXT NOT NULL DEFAULT 'sAMAccountName',
    email_attr        TEXT NOT NULL DEFAULT 'mail',
    tls_mode          TEXT NOT NULL DEFAULT 'none' CHECK (tls_mode IN ('none', 'starttls', 'tls')),
    tls_skip_verify   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);