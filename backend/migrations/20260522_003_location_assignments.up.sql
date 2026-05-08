-- +migrate Up

ALTER TABLE subnets ADD COLUMN location_id INT REFERENCES locations(id) ON DELETE SET NULL;
ALTER TABLE devices ADD COLUMN location_id INT REFERENCES locations(id) ON DELETE SET NULL;
