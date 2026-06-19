-- +migrate Up
CREATE TABLE role_grants (
  id              BIGSERIAL    PRIMARY KEY,
  organization_id BIGINT       NOT NULL REFERENCES organizations(id),
  user_id         BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  permission      TEXT         NOT NULL,
  scope_type      TEXT,
  scope_id        BIGINT,
  granted_by      BIGINT       REFERENCES users(id) ON DELETE SET NULL,
  granted_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX role_grants_user_perm ON role_grants (user_id, permission);
