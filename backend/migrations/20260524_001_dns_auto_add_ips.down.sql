-- +migrate Down
DELETE FROM configs WHERE key IN ('dns_auto_add_ips_enabled', 'dns_auto_remove_ips_enabled');
