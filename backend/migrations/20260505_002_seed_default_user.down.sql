-- +migrate Down

-- Remove default user
DELETE FROM users WHERE username = 'admin' AND email = 'admin@localhost';
