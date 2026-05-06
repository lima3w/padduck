-- +migrate Down

DELETE FROM config WHERE key IN ('session_idle_timeout_minutes', 'session_absolute_timeout_hours');

DROP TABLE IF EXISTS sessions;
