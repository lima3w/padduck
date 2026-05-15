CREATE TABLE ipv6_delegations (
    id BIGSERIAL PRIMARY KEY,
    parent_subnet_id BIGINT NOT NULL REFERENCES subnets(id) ON DELETE CASCADE,
    delegated_prefix VARCHAR(50) NOT NULL,
    delegated_to_device_id BIGINT REFERENCES devices(id) ON DELETE SET NULL,
    delegated_to_description TEXT,
    valid_lifetime_sec INTEGER,
    preferred_lifetime_sec INTEGER,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
