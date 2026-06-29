-- +migrate Up
ALTER TABLE automation_policies ADD COLUMN actions JSONB NOT NULL DEFAULT '[]';
