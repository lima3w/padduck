-- +migrate Up
ALTER TABLE api_tokens ADD COLUMN bypass_policy BOOLEAN NOT NULL DEFAULT false;
