-- +migrate Down

DELETE FROM config WHERE key = 'privacy_policy_version';

ALTER TABLE sessions
    DROP COLUMN IF EXISTS is_impersonation,
    DROP COLUMN IF EXISTS impersonated_by;

ALTER TABLE users
    DROP COLUMN IF EXISTS deletion_requested_at,
    DROP COLUMN IF EXISTS anonymized_at,
    DROP COLUMN IF EXISTS privacy_accepted_at,
    DROP COLUMN IF EXISTS privacy_accepted_version,
    DROP COLUMN IF EXISTS suspended_at,
    DROP COLUMN IF EXISTS suspended_by,
    DROP COLUMN IF EXISTS suspension_reason;
