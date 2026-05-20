-- +migrate Up
CREATE TABLE custom_field_definitions (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('subnet', 'ip_address', 'device')),
    name VARCHAR(100) NOT NULL,
    label VARCHAR(200) NOT NULL,
    field_type VARCHAR(20) NOT NULL CHECK (field_type IN ('text','number','textarea','dropdown','checkbox','date','url','email')),
    options JSONB,
    is_required BOOLEAN NOT NULL DEFAULT false,
    default_value VARCHAR(500),
    placeholder VARCHAR(200),
    display_order INT NOT NULL DEFAULT 0,
    is_searchable BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE custom_field_values (
    id SERIAL PRIMARY KEY,
    definition_id INT NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    entity_id INT NOT NULL,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('subnet', 'ip_address', 'device')),
    value TEXT,
    UNIQUE(definition_id, entity_id, entity_type)
);

CREATE INDEX idx_cfv_entity ON custom_field_values(entity_type, entity_id);
CREATE INDEX idx_cfv_definition ON custom_field_values(definition_id);
