-- +migrate Up

-- Make created_by nullable in sections table
ALTER TABLE sections
ALTER COLUMN created_by DROP NOT NULL;
