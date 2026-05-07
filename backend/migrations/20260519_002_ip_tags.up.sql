-- +migrate Up

CREATE TABLE ip_tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    colour VARCHAR(7) NOT NULL DEFAULT '#6B7280',
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO ip_tags (name, colour, description, is_system) VALUES
    ('Used',     '#22C55E', 'IP is actively assigned',         true),
    ('Free',     '#F9FAFB', 'IP is available',                 true),
    ('Reserved', '#9CA3AF', 'IP is administratively reserved', true),
    ('DHCP',     '#3B82F6', 'IP managed by DHCP',              true),
    ('Offline',  '#EF4444', 'IP is not responding',            true);

ALTER TABLE ip_addresses ADD COLUMN tag_id BIGINT REFERENCES ip_tags(id) ON DELETE SET NULL;
