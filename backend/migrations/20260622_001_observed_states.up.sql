-- +migrate Up
CREATE TABLE observed_states (
  id              BIGSERIAL    PRIMARY KEY,
  organization_id BIGINT       REFERENCES organizations(id) ON DELETE CASCADE,
  resource_type   TEXT         NOT NULL,
  resource_id     BIGINT,
  ip_address      TEXT,
  observed_data   JSONB        NOT NULL DEFAULT '{}',
  source          TEXT         NOT NULL DEFAULT 'scan',
  scan_result_id  BIGINT       REFERENCES scan_results(id) ON DELETE SET NULL,
  first_seen_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  last_seen_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Unique per registered resource (resource_id is the authoritative record FK).
CREATE UNIQUE INDEX observed_states_resource_udx
  ON observed_states(resource_type, resource_id)
  WHERE resource_id IS NOT NULL;

-- Unique per unregistered IP seen by scanner.
CREATE UNIQUE INDEX observed_states_unregistered_udx
  ON observed_states(ip_address)
  WHERE resource_id IS NULL AND ip_address IS NOT NULL;
