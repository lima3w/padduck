-- +migrate Down
DROP TABLE IF EXISTS request_comments;
DROP TYPE IF EXISTS request_comment_type;
