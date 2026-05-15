-- +migrate Up

ALTER TABLE scan_results ADD COLUMN ptr_record TEXT;
ALTER TABLE scan_results ADD COLUMN fwd_rev_mismatch BOOLEAN NOT NULL DEFAULT FALSE;
