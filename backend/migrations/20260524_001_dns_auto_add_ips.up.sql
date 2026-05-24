-- +migrate Up
INSERT INTO configs (key, value) VALUES ('dns_auto_add_ips_enabled', 'false') ON CONFLICT (key) DO NOTHING;
INSERT INTO configs (key, value) VALUES ('dns_auto_remove_ips_enabled', 'false') ON CONFLICT (key) DO NOTHING;
