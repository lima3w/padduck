-- +migrate Up
-- resource_intents.organization_id was nullable and missed by the organization
-- isolation migration. Backfill from the submitter, then require it.
UPDATE resource_intents i SET organization_id = u.organization_id
  FROM users u WHERE i.submitted_by = u.id AND i.organization_id IS NULL;
UPDATE resource_intents SET organization_id = 1 WHERE organization_id IS NULL;
ALTER TABLE resource_intents ALTER COLUMN organization_id SET DEFAULT 1;
ALTER TABLE resource_intents ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS resource_intents_org_status_idx
  ON resource_intents(organization_id, status);
