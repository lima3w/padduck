-- +migrate Down

DELETE FROM configs
WHERE key IN (
	'feature_customers_enabled',
	'feature_vlans_enabled',
	'feature_vrfs_enabled',
	'feature_racks_enabled',
	'feature_locations_enabled',
	'feature_bgp_enabled',
	'feature_devices_enabled'
);
