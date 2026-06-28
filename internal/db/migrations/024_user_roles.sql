-- +goose Up

-- user_roles: a user may hold several roles; their effective permissions are
-- the union across all of them. users.role_id is kept as a denormalised
-- "primary" role for single-role display and backward compatibility.
CREATE TABLE user_roles (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- Seed the junction from the existing single role assignment.
INSERT INTO user_roles (user_id, role_id)
SELECT id, role_id FROM users WHERE role_id IS NOT NULL;

-- +goose Down
DROP TABLE user_roles;
