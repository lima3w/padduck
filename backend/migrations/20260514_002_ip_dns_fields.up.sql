-- +migrate Up

ALTER TABLE ip_addresses
    ADD COLUMN dns_name         VARCHAR(253),
    ADD COLUMN dns_records      JSONB,
    ADD COLUMN dns_last_checked TIMESTAMPTZ;
