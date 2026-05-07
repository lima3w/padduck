-- +migrate Up
INSERT INTO configs (key, value) VALUES
    ('api_token_default_expiration_days', '30'),
    ('api_token_rotation_grace_period_hours', '24'),
    ('api_token_rate_limit_per_minute', '100')
ON CONFLICT (key) DO NOTHING;
