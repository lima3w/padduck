-- +migrate Down

ALTER TABLE scan_results DROP COLUMN IF EXISTS fwd_rev_mismatch;
ALTER TABLE scan_results DROP COLUMN IF EXISTS ptr_record;
