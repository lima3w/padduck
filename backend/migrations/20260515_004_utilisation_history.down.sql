-- +migrate Down
DROP INDEX IF EXISTS idx_util_history_subnet_time;
DROP TABLE IF EXISTS subnet_utilisation_history;
