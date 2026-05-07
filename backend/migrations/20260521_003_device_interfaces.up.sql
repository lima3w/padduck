-- +migrate Up
CREATE TABLE device_interfaces (
  id SERIAL PRIMARY KEY,
  device_id INT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  speed_mbps INT,
  media_type VARCHAR(50),
  vlan_id INT REFERENCES vlans(id) ON DELETE SET NULL,
  ip_address_id INT REFERENCES ip_addresses(id) ON DELETE SET NULL,
  connected_to_device_id INT REFERENCES devices(id) ON DELETE SET NULL,
  connected_to_interface_id INT REFERENCES device_interfaces(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
