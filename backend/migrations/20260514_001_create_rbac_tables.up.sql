-- +migrate Up

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Partial unique indexes handle NULLs correctly in PostgreSQL
CREATE UNIQUE INDEX idx_rp_global
    ON role_permissions(role_id, permission)
    WHERE resource_type IS NULL AND resource_id IS NULL;

CREATE UNIQUE INDEX idx_rp_typed
    ON role_permissions(role_id, permission, resource_type)
    WHERE resource_id IS NULL AND resource_type IS NOT NULL;

CREATE UNIQUE INDEX idx_rp_specific
    ON role_permissions(role_id, permission, resource_type, resource_id)
    WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
