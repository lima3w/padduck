-- +migrate Up
CREATE TABLE resource_intents (
  id              BIGSERIAL    PRIMARY KEY,
  organization_id BIGINT       REFERENCES organizations(id) ON DELETE CASCADE,
  resource_type   TEXT         NOT NULL,
  resource_id     BIGINT,
  operation       TEXT         NOT NULL,
  desired_state   JSONB        NOT NULL DEFAULT '{}',
  status          TEXT         NOT NULL DEFAULT 'pending',
  submitted_by    BIGINT       REFERENCES users(id) ON DELETE SET NULL,
  reviewed_by     BIGINT       REFERENCES users(id) ON DELETE SET NULL,
  reviewer_note   TEXT,
  submitted_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  reviewed_at     TIMESTAMPTZ,
  applied_at      TIMESTAMPTZ
);

CREATE INDEX resource_intents_resource_idx
  ON resource_intents(resource_type, resource_id)
  WHERE resource_id IS NOT NULL;

CREATE INDEX resource_intents_status_idx
  ON resource_intents(status);
