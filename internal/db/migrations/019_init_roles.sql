-- +goose Up

-- roles: named bundles of instance permissions. System roles are seeded and
-- cannot be deleted; custom roles are created from the admin UI.
CREATE TABLE roles (
    id         TEXT    NOT NULL PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE COLLATE NOCASE,
    is_system  INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE role_permissions (
    role_id    TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    PRIMARY KEY (role_id, permission)
);

-- users reference a role. Nullable: a NULL role_id means the base "user" role
-- (no instance permissions). The legacy users.role column is kept in sync as a
-- coarse user/admin fallback that still satisfies its CHECK constraint.
ALTER TABLE users ADD COLUMN role_id TEXT REFERENCES roles(id);

-- Seed the three system roles with stable IDs (referenced from Go).
INSERT INTO roles (id, name, is_system, created_at, updated_at) VALUES
    ('role_owner', 'owner', 1, CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('role_admin', 'admin', 1, CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)),
    ('role_user',  'user',  1, CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER));

-- owner holds every permission.
INSERT INTO role_permissions (role_id, permission) VALUES
    ('role_owner', 'audit.view'),
    ('role_owner', 'bookmarks.import_export'),
    ('role_owner', 'collections.create'),
    ('role_owner', 'collections.manage_all'),
    ('role_owner', 'collections.delete_any'),
    ('role_owner', 'groups.manage'),
    ('role_owner', 'users.manage'),
    ('role_owner', 'settings.manage'),
    ('role_owner', 'roles.manage');

-- admin: day-to-day administration, minus role/policy super-powers.
INSERT INTO role_permissions (role_id, permission) VALUES
    ('role_admin', 'audit.view'),
    ('role_admin', 'collections.create'),
    ('role_admin', 'groups.manage'),
    ('role_admin', 'users.manage'),
    ('role_admin', 'settings.manage');

-- Backfill: map the legacy role column onto role_id.
UPDATE users SET role_id = 'role_user'  WHERE role = 'user';
UPDATE users SET role_id = 'role_admin' WHERE role = 'admin';

-- The very first user (earliest created) becomes the instance owner.
UPDATE users SET role_id = 'role_owner'
WHERE id = (SELECT id FROM users ORDER BY created_at ASC, id ASC LIMIT 1);

CREATE INDEX idx_users_role_id ON users(role_id);

-- +goose Down
DROP INDEX IF EXISTS idx_users_role_id;
ALTER TABLE users DROP COLUMN role_id;
DROP TABLE role_permissions;
DROP TABLE roles;
