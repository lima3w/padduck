-- +migrate Up
-- #202 Subnet Request Workflow
CREATE TYPE subnet_request_status AS ENUM ('pending', 'approved', 'rejected', 'cancelled');

CREATE TABLE subnet_requests (
    id                   BIGSERIAL PRIMARY KEY,
    requester_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    section_id           BIGINT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    parent_subnet_id     BIGINT REFERENCES subnets(id) ON DELETE SET NULL,
    requested_prefix_len INT NOT NULL,
    purpose              TEXT NOT NULL DEFAULT '',
    status               subnet_request_status NOT NULL DEFAULT 'pending',
    reviewer_id          BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewer_note        TEXT NOT NULL DEFAULT '',
    subnet_id            BIGINT REFERENCES subnets(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subnet_requests_requester ON subnet_requests(requester_id);
CREATE INDEX idx_subnet_requests_status ON subnet_requests(status);
CREATE INDEX idx_subnet_requests_section ON subnet_requests(section_id);
