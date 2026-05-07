-- +migrate Up
CREATE TABLE device_types (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  icon VARCHAR(50),
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO device_types (name, icon) VALUES
  ('Router', 'router'),
  ('Switch', 'switch'),
  ('Firewall', 'firewall'),
  ('Server', 'server'),
  ('Access Point', 'wifi'),
  ('Load Balancer', 'balance'),
  ('Storage', 'storage'),
  ('Virtual Machine', 'vm'),
  ('Printer', 'printer'),
  ('Phone', 'phone');

CREATE TABLE devices (
  id SERIAL PRIMARY KEY,
  hostname VARCHAR(255) NOT NULL,
  description TEXT,
  type_id INT REFERENCES device_types(id) ON DELETE SET NULL,
  section_id INT REFERENCES sections(id) ON DELETE SET NULL,
  vendor VARCHAR(100),
  model VARCHAR(100),
  os_version VARCHAR(100),
  snmp_community VARCHAR(500),
  snmp_version VARCHAR(10) DEFAULT 'v2c',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
