-- +migrate Up

ALTER TABLE ip_addresses
ADD COLUMN assigned_at TIMESTAMP NULL,
ADD COLUMN expires_at TIMESTAMP NULL;

-- Set assigned_at for already assigned IPs (use updated_at as approximation)
UPDATE ip_addresses
SET assigned_at = updated_at
WHERE status = 'assigned' AND assigned_to IS NOT NULL;
