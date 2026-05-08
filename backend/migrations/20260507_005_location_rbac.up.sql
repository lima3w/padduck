-- +migrate Up

ALTER TABLE user_roles ADD COLUMN location_id INT REFERENCES locations(id) ON DELETE CASCADE;
