-- +migrate Up

CREATE TABLE racks (
  id SERIAL PRIMARY KEY,
  location_id INT REFERENCES locations(id) ON DELETE SET NULL,
  name VARCHAR(255) NOT NULL,
  size_u INT NOT NULL DEFAULT 42,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE devices ADD COLUMN rack_id INT REFERENCES racks(id) ON DELETE SET NULL;
ALTER TABLE devices ADD COLUMN rack_unit_start INT;
ALTER TABLE devices ADD COLUMN rack_unit_size INT NOT NULL DEFAULT 1;
