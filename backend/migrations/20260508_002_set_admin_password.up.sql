-- +migrate Up

-- Leave password_hash NULL so the app sets it on first boot from ADMIN_PASSWORD
-- env var or a generated random password printed to stdout.
UPDATE users
SET password_hash = NULL
WHERE username = 'admin';
