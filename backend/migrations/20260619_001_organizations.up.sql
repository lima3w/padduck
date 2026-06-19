-- +migrate Up
CREATE TABLE organizations (
  id         BIGSERIAL    PRIMARY KEY,
  name       TEXT         NOT NULL,
  slug       TEXT         NOT NULL UNIQUE,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN organization_id BIGINT REFERENCES organizations(id);

-- Seed: create the default organization and assign all existing users to it.
INSERT INTO organizations (id, name, slug) VALUES (1, 'Default', 'default');
SELECT setval('organizations_id_seq', 1);
UPDATE users SET organization_id = 1 WHERE organization_id IS NULL;
