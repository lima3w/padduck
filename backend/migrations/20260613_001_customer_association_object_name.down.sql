-- +migrate Down
ALTER TABLE customer_associations DROP COLUMN IF EXISTS object_name;
