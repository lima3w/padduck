-- +migrate Up
-- Migrate existing users to the new role system based on their legacy role column.
-- role=admin → admin role, role=user → operator role, role=viewer → viewer role
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON (
    (u.role = 'admin'  AND r.name = 'admin')   OR
    (u.role = 'user'   AND r.name = 'operator') OR
    (u.role = 'viewer' AND r.name = 'viewer')
)
ON CONFLICT DO NOTHING;
