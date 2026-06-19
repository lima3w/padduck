CREATE TABLE background_jobs (
  id            BIGSERIAL    PRIMARY KEY,
  type          TEXT         NOT NULL,
  name          TEXT         NOT NULL DEFAULT '',
  status        TEXT         NOT NULL DEFAULT 'queued',
  progress      INT          NOT NULL DEFAULT 0,
  diagnostic    TEXT         NOT NULL DEFAULT '',
  error_message TEXT         NOT NULL DEFAULT '',
  payload       JSONB,
  result        JSONB,
  attempts      INT          NOT NULL DEFAULT 0,
  max_attempts  INT          NOT NULL DEFAULT 1,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  started_at    TIMESTAMPTZ,
  finished_at   TIMESTAMPTZ
);

CREATE INDEX background_jobs_status_created ON background_jobs (status, created_at DESC);
