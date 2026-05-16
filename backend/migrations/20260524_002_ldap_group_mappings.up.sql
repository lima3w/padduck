-- +migrate Up

CREATE TABLE IF NOT EXISTS ldap_group_role_mappings (
    id            BIGSERIAL PRIMARY KEY,
    ldap_group_dn TEXT NOT NULL,
    role_id       BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ldap_group_dn)
);