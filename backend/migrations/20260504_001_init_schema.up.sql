-- +migrate Up

-- Create initial schema for IPAM Next

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sections (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subnets (
    id BIGSERIAL PRIMARY KEY,
    section_id BIGINT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    network_address INET NOT NULL,
    prefix_length INTEGER NOT NULL CHECK (prefix_length >= 0 AND prefix_length <= 32),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(section_id, network_address, prefix_length)
);

CREATE TABLE IF NOT EXISTS ip_addresses (
    id BIGSERIAL PRIMARY KEY,
    subnet_id BIGINT NOT NULL REFERENCES subnets(id) ON DELETE CASCADE,
    address INET NOT NULL,
    hostname VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'assigned', 'reserved')),
    assigned_to VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(subnet_id, address)
);

-- Create indexes for common queries
CREATE INDEX idx_sections_created_by ON sections(created_by);
CREATE INDEX idx_subnets_section_id ON subnets(section_id);
CREATE INDEX idx_ip_addresses_subnet_id ON ip_addresses(subnet_id);
CREATE INDEX idx_ip_addresses_status ON ip_addresses(status);
CREATE INDEX idx_ip_addresses_assigned_to ON ip_addresses(assigned_to);
