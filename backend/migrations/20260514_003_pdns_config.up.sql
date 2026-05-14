-- +migrate Up

INSERT INTO configs (key, value) VALUES
    ('pdns_enabled',      'false'),
    ('pdns_api_url',      ''),
    ('pdns_api_key',      ''),
    ('pdns_default_zone', ''),
    ('pdns_ptr_zones',    '')
ON CONFLICT (key) DO NOTHING;
