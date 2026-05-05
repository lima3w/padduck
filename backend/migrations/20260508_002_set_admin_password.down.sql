-- +migrate Down

-- Clear password hash
UPDATE users
SET password_hash = NULL
WHERE username = 'admin';
