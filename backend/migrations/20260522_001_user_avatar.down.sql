-- +migrate Down

ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_data,
    DROP COLUMN IF EXISTS avatar_source;
