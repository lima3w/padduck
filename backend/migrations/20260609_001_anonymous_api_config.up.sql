INSERT INTO configs (key, value) VALUES ('anonymous_api_enabled', 'false') ON CONFLICT (key) DO NOTHING;
