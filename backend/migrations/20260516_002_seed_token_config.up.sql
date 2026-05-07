-- +migrate Up
INSERT INTO configs (key, value, description) VALUES
    ('api_token_default_expiration_days', '30', 'Default token lifetime in days (0 = no expiration)'),
    ('api_token_rotation_grace_period_hours', '24', 'Hours old token remains valid after rotation'),
    ('api_token_rate_limit_per_minute', '100', 'Max requests per minute per API token (0 = unlimited)')
ON CONFLICT (key) DO NOTHING;
