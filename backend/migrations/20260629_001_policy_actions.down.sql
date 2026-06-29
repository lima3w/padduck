-- +migrate Down
ALTER TABLE automation_policies DROP COLUMN IF EXISTS actions;
