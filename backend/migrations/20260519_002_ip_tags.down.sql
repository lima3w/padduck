-- +migrate Down

ALTER TABLE ip_addresses DROP COLUMN IF EXISTS tag_id;
DROP TABLE IF EXISTS ip_tags;
