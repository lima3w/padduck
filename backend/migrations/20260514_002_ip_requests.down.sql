-- +migrate Down
DROP TABLE IF EXISTS ip_requests;
DROP TYPE IF EXISTS ip_request_status;
