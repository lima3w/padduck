-- +migrate Down
DELETE FROM configs WHERE key IN (
    'api_token_default_expiration_days',
    'api_token_rotation_grace_period_hours',
    'api_token_rate_limit_per_minute'
);
