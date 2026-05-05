-- +migrate Up

-- Set password for default admin user (bcrypt hash of "admin")
-- This should be changed in production
UPDATE users
SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoHyeiYlXSyaWjrNmPIb7Dbt0HQzcdmBVWy6'
WHERE username = 'admin';
