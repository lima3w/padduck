-- +migrate Up
CREATE TABLE topology_nodes (
    id              BIGSERIAL    PRIMARY KEY,
    organization_id BIGINT       REFERENCES organizations(id) ON DELETE CASCADE,
    resource_type   TEXT         NOT NULL,
    resource_id     BIGINT       NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (resource_type, resource_id)
);

CREATE TABLE topology_edges (
    id              BIGSERIAL    PRIMARY KEY,
    organization_id BIGINT       REFERENCES organizations(id) ON DELETE CASCADE,
    source_node_id  BIGINT       NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    target_node_id  BIGINT       NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    relationship    TEXT         NOT NULL,
    confidence      FLOAT        NOT NULL DEFAULT 1.0,
    status          TEXT         NOT NULL DEFAULT 'confirmed',
    evidence        TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (source_node_id, target_node_id, relationship)
);

CREATE INDEX topology_edges_source_idx ON topology_edges(source_node_id);
CREATE INDEX topology_edges_target_idx ON topology_edges(target_node_id);
