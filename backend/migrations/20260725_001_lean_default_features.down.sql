-- +migrate Down
-- Data-only change to defaults; nothing to undo. Whether a module is on is an
-- admin choice and is not reverted.
SELECT 1;
