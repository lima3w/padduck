-- +migrate Down

ALTER TABLE ip_addresses
    DROP COLUMN IF EXISTS dns_last_checked,
    DROP COLUMN IF EXISTS dns_records,
    DROP COLUMN IF EXISTS dns_name;
