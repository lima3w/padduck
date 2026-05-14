-- +migrate Down
DROP TABLE IF EXISTS subnet_requests;
DROP TYPE IF EXISTS subnet_request_status;
