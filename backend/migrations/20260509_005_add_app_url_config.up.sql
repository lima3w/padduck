-- +migrate Up

INSERT INTO configs (key, value) VALUES ('app_url', 'http://localhost:3000')
ON CONFLICT (key) DO NOTHING;
