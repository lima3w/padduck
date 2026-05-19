-- +migrate Down
ALTER TABLE subnets DROP COLUMN IF EXISTS scan_profile_id;
DROP TABLE IF EXISTS scan_profiles;
