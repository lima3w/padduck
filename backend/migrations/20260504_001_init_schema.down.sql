-- +migrate Down

-- Rollback initial schema

DROP TABLE IF EXISTS ip_addresses;
DROP TABLE IF EXISTS subnets;
DROP TABLE IF EXISTS sections;
DROP TABLE IF EXISTS users;
