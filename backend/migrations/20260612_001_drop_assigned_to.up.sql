-- +migrate Up
DROP INDEX IF EXISTS idx_ip_addresses_assigned_to;
ALTER TABLE ip_addresses DROP COLUMN IF EXISTS assigned_to;
