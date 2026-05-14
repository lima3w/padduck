-- +migrate Up
-- #203 IP Address Request Workflow
CREATE TYPE ip_request_status AS ENUM ('pending', 'approved', 'rejected', 'cancelled');

CREATE TABLE ip_requests (
    id              BIGSERIAL PRIMARY KEY,
    requester_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subnet_id       BIGINT NOT NULL REFERENCES subnets(id) ON DELETE CASCADE,
    requested_ip    VARCHAR(45),
    dns_name        TEXT NOT NULL DEFAULT '',
    purpose         TEXT NOT NULL DEFAULT '',
    status          ip_request_status NOT NULL DEFAULT 'pending',
    reviewer_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewer_note   TEXT NOT NULL DEFAULT '',
    ip_address_id   BIGINT REFERENCES ip_addresses(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ip_requests_requester ON ip_requests(requester_id);
CREATE INDEX idx_ip_requests_status ON ip_requests(status);
CREATE INDEX idx_ip_requests_subnet ON ip_requests(subnet_id);
