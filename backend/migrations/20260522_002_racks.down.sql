-- +migrate Down

ALTER TABLE devices DROP COLUMN IF EXISTS rack_unit_size;
ALTER TABLE devices DROP COLUMN IF EXISTS rack_unit_start;
ALTER TABLE devices DROP COLUMN IF EXISTS rack_id;
DROP TABLE IF EXISTS racks;
