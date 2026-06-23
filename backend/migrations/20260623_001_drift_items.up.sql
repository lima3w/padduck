-- +migrate Up
CREATE TABLE drift_items (
  id                BIGSERIAL    PRIMARY KEY,
  organization_id   BIGINT       REFERENCES organizations(id) ON DELETE CASCADE,
  resource_type     TEXT         NOT NULL,
  resource_id       BIGINT       NOT NULL,
  observed_state_id BIGINT       NOT NULL REFERENCES observed_states(id) ON DELETE CASCADE,
  field_diffs       JSONB        NOT NULL,
  status            TEXT         NOT NULL DEFAULT 'open',
  resolved_by       BIGINT       REFERENCES users(id),
  resolved_at       TIMESTAMPTZ,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- At most one open drift item per resource at a time.
CREATE UNIQUE INDEX drift_items_open_udx
  ON drift_items(resource_type, resource_id)
  WHERE status = 'open';
