-- +migrate Up

ALTER TABLE subnets ADD COLUMN gateway VARCHAR(45);
ALTER TABLE subnets ADD COLUMN auto_reserve_first BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE subnets ADD COLUMN auto_reserve_last BOOLEAN NOT NULL DEFAULT false;
