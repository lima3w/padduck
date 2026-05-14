-- +migrate Down

DELETE FROM configs WHERE key IN (
    'pdns_enabled',
    'pdns_api_url',
    'pdns_api_key',
    'pdns_default_zone',
    'pdns_ptr_zones'
);
