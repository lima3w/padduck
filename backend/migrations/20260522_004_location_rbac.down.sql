-- +migrate Down

ALTER TABLE user_roles DROP COLUMN IF EXISTS location_id;
