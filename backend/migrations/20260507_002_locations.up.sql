-- +migrate Up

CREATE TABLE locations (
  id SERIAL PRIMARY KEY,
  parent_id INT REFERENCES locations(id) ON DELETE SET NULL,
  name VARCHAR(255) NOT NULL,
  type VARCHAR(50) NOT NULL DEFAULT 'other',
  address TEXT,
  lat DECIMAL(10,7),
  lng DECIMAL(10,7),
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
