package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type RoleRepository interface {
	ListAll(ctx context.Context) ([]*model.Role, error)
	GetByID(ctx context.Context, id string) (*model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id string) error
	SetPermissions(ctx context.Context, roleID string, perms []string) error
	// PermissionsForUser returns the union of permissions across all the user's roles.
	PermissionsForUser(ctx context.Context, userID string) ([]string, error)
	// RolesForUser returns every role held by the user.
	RolesForUser(ctx context.Context, userID string) ([]model.Role, error)
	// RolesForUsers batch-loads roles for several users (admin list).
	RolesForUsers(ctx context.Context, userIDs []string) (map[string][]model.Role, error)
	// SetUserRoles replaces the user's full role set; primaryRoleID/legacyRole keep
	// users.role_id and users.role (the coarse fallback) in sync.
	SetUserRoles(ctx context.Context, userID string, roleIDs []string, primaryRoleID, legacyRole string) error
	// AssignUserRole sets a user's single role (convenience wrapper over SetUserRoles).
	AssignUserRole(ctx context.Context, userID, roleID, legacyRole string) error
	// AddUserRole adds one role to a user (and seeds users.role_id if empty).
	AddUserRole(ctx context.Context, userID, roleID, legacyRole string) error
	// CountUsersWithRole counts active users currently holding a role.
	CountUsersWithRole(ctx context.Context, roleID string) (int, error)
}

type sqliteRoleRepo struct {
	db *sql.DB
}

func newSQLiteRoleRepo(db *sql.DB) RoleRepository {
	return &sqliteRoleRepo{db: db}
}

func (r *sqliteRoleRepo) ListAll(ctx context.Context) ([]*model.Role, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, is_system, created_at, updated_at FROM roles ORDER BY is_system DESC, name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, fmt.Errorf("role list: %w", err)
	}
	defer rows.Close()

	var roles []*model.Role
	byID := map[string]*model.Role{}
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
		byID[role.ID] = role
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Attach permissions in a single sweep.
	permRows, err := r.db.QueryContext(ctx, `SELECT role_id, permission FROM role_permissions`)
	if err != nil {
		return nil, fmt.Errorf("role perms: %w", err)
	}
	defer permRows.Close()
	for permRows.Next() {
		var roleID, perm string
		if err := permRows.Scan(&roleID, &perm); err != nil {
			return nil, err
		}
		if role := byID[roleID]; role != nil {
			role.Permissions = append(role.Permissions, perm)
		}
	}
	return roles, permRows.Err()
}

func (r *sqliteRoleRepo) GetByID(ctx context.Context, id string) (*model.Role, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, is_system, created_at, updated_at FROM roles WHERE id = ?`, id)
	role, err := scanRole(row)
	if err != nil {
		return nil, err
	}
	role.Permissions, err = r.permsForRole(ctx, id)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *sqliteRoleRepo) GetByName(ctx context.Context, name string) (*model.Role, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, is_system, created_at, updated_at FROM roles WHERE name = ? COLLATE NOCASE`, name)
	role, err := scanRole(row)
	if err != nil {
		return nil, err
	}
	role.Permissions, err = r.permsForRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *sqliteRoleRepo) Create(ctx context.Context, role *model.Role) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO roles (id, name, is_system, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		role.ID, role.Name, boolInt(role.IsSystem), role.CreatedAt.Unix(), role.UpdatedAt.Unix(),
	); err != nil {
		return fmt.Errorf("role create: %w", err)
	}
	if err := insertPerms(ctx, tx, role.ID, role.Permissions); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *sqliteRoleRepo) Update(ctx context.Context, role *model.Role) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE roles SET name = ?, updated_at = ? WHERE id = ?`,
		role.Name, time.Now().Unix(), role.ID,
	); err != nil {
		return fmt.Errorf("role update: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, role.ID); err != nil {
		return err
	}
	if err := insertPerms(ctx, tx, role.ID, role.Permissions); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *sqliteRoleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = ?`, id)
	return err
}

func (r *sqliteRoleRepo) SetPermissions(ctx context.Context, roleID string, perms []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, roleID); err != nil {
		return err
	}
	if err := insertPerms(ctx, tx, roleID, perms); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *sqliteRoleRepo) PermissionsForUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT rp.permission
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("perms for user: %w", err)
	}
	defer rows.Close()
	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *sqliteRoleRepo) RolesForUser(ctx context.Context, userID string) ([]model.Role, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT rl.id, rl.name, rl.is_system, rl.created_at, rl.updated_at
		FROM user_roles ur JOIN roles rl ON rl.id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY rl.is_system DESC, rl.name COLLATE NOCASE ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("roles for user: %w", err)
	}
	defer rows.Close()
	var out []model.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *role)
	}
	return out, rows.Err()
}

func (r *sqliteRoleRepo) RolesForUsers(ctx context.Context, userIDs []string) (map[string][]model.Role, error) {
	out := map[string][]model.Role{}
	if len(userIDs) == 0 {
		return out, nil
	}
	placeholders := strings.Repeat("?,", len(userIDs)-1) + "?"
	args := make([]any, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT ur.user_id, rl.id, rl.name, rl.is_system, rl.created_at, rl.updated_at
		FROM user_roles ur JOIN roles rl ON rl.id = ur.role_id
		WHERE ur.user_id IN (`+placeholders+`)
		ORDER BY rl.is_system DESC, rl.name COLLATE NOCASE ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("roles for users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var uid string
		var role model.Role
		var isSystem int
		var createdAt, updatedAt int64
		if err := rows.Scan(&uid, &role.ID, &role.Name, &isSystem, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		role.IsSystem = isSystem != 0
		role.CreatedAt = time.Unix(createdAt, 0)
		role.UpdatedAt = time.Unix(updatedAt, 0)
		out[uid] = append(out[uid], role)
	}
	return out, rows.Err()
}

func (r *sqliteRoleRepo) SetUserRoles(ctx context.Context, userID string, roleIDs []string, primaryRoleID, legacyRole string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("clear user roles: %w", err)
	}
	for _, rid := range roleIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, rid,
		); err != nil {
			return fmt.Errorf("add user role: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET role_id = ?, role = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		nullStr(primaryRoleID), legacyRole, time.Now().Unix(), userID,
	); err != nil {
		return fmt.Errorf("update primary role: %w", err)
	}
	return tx.Commit()
}

func (r *sqliteRoleRepo) AssignUserRole(ctx context.Context, userID, roleID, legacyRole string) error {
	return r.SetUserRoles(ctx, userID, []string{roleID}, roleID, legacyRole)
}

func (r *sqliteRoleRepo) AddUserRole(ctx context.Context, userID, roleID, legacyRole string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, roleID,
	); err != nil {
		return fmt.Errorf("add user role: %w", err)
	}
	// Seed the primary role if none set yet.
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET role_id = COALESCE(NULLIF(role_id, ''), ?), role = ?, updated_at = ?
		 WHERE id = ? AND deleted_at IS NULL`,
		roleID, legacyRole, time.Now().Unix(), userID,
	); err != nil {
		return fmt.Errorf("seed primary role: %w", err)
	}
	return tx.Commit()
}

func (r *sqliteRoleRepo) CountUsersWithRole(ctx context.Context, roleID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT ur.user_id) FROM user_roles ur
		JOIN users u ON u.id = ur.user_id
		WHERE ur.role_id = ? AND u.deleted_at IS NULL`, roleID).Scan(&n)
	return n, err
}

func (r *sqliteRoleRepo) permsForRole(ctx context.Context, roleID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT permission FROM role_permissions WHERE role_id = ?`, roleID)
	if err != nil {
		return nil, fmt.Errorf("perms for role: %w", err)
	}
	defer rows.Close()
	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func insertPerms(ctx context.Context, tx *sql.Tx, roleID string, perms []string) error {
	for _, p := range perms {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO role_permissions (role_id, permission) VALUES (?, ?)`, roleID, p,
		); err != nil {
			return fmt.Errorf("insert perm: %w", err)
		}
	}
	return nil
}

func scanRole(s scanner) (*model.Role, error) {
	var role model.Role
	var isSystem int
	var createdAt, updatedAt int64
	if err := s.Scan(&role.ID, &role.Name, &isSystem, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan role: %w", err)
	}
	role.IsSystem = isSystem != 0
	role.CreatedAt = time.Unix(createdAt, 0)
	role.UpdatedAt = time.Unix(updatedAt, 0)
	return &role, nil
}
