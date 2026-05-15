-- +migrate Up

CREATE TABLE vlan_groups (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR NOT NULL UNIQUE,
    description TEXT,
    colour      VARCHAR(7),
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE vlans ADD COLUMN group_id INT REFERENCES vlan_groups(id) ON DELETE SET NULL;

CREATE INDEX idx_vlans_group_id ON vlans(group_id);
