-- +migrate Up
ALTER TABLE subnets DROP CONSTRAINT IF EXISTS subnets_prefix_length_check;
ALTER TABLE subnets ADD CONSTRAINT subnets_prefix_length_check CHECK (prefix_length >= 0 AND prefix_length <= 128);
