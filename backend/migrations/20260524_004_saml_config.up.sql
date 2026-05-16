-- +migrate Up

CREATE TABLE IF NOT EXISTS saml_configs (
    id               BIGSERIAL PRIMARY KEY,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    idp_metadata_url TEXT NOT NULL DEFAULT '',
    idp_metadata_xml TEXT NOT NULL DEFAULT '',
    sp_cert_pem      TEXT NOT NULL DEFAULT '',
    sp_key_pem       TEXT NOT NULL DEFAULT '',
    entity_id        TEXT NOT NULL DEFAULT '',
    acs_url          TEXT NOT NULL DEFAULT '',
    name_id_format   TEXT NOT NULL DEFAULT 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);