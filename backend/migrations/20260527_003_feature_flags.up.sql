-- +migrate Up

INSERT INTO configs (key, value) VALUES
	('feature_customers_enabled', 'true'),
	('feature_vlans_enabled', 'true'),
	('feature_vrfs_enabled', 'true'),
	('feature_racks_enabled', 'true'),
	('feature_locations_enabled', 'true'),
	('feature_bgp_enabled', 'true'),
	('feature_devices_enabled', 'true')
ON CONFLICT (key) DO NOTHING;
