-- +migrate Down

-- Revert created_by to NOT NULL
ALTER TABLE sections
ALTER COLUMN created_by SET NOT NULL;
