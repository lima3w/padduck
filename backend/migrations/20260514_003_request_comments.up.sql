-- +migrate Up
-- #204 Request Comments and Audit Trail
CREATE TYPE request_comment_type AS ENUM ('subnet', 'ip');

CREATE TABLE request_comments (
    id           BIGSERIAL PRIMARY KEY,
    request_type request_comment_type NOT NULL,
    request_id   BIGINT NOT NULL,
    author_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body         TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_comments_request ON request_comments(request_type, request_id);
CREATE INDEX idx_request_comments_author ON request_comments(author_id);
