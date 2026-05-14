-- +migrate Up

CREATE TABLE nameservers (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    server1     VARCHAR(253) NOT NULL,
    server2     VARCHAR(253),
    server3     VARCHAR(253),
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE subnets
    ADD COLUMN nameserver_id INT REFERENCES nameservers(id) ON DELETE SET NULL;
