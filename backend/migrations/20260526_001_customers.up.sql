-- +migrate Up
CREATE TABLE customers (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
