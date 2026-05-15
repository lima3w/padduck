-- +migrate Up

CREATE TABLE vlan_domains (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE vlans ADD COLUMN domain_id INT REFERENCES vlan_domains(id) ON DELETE SET NULL;

CREATE INDEX idx_vlans_domain_id ON vlans(domain_id);
