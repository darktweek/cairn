package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	// PermissionsForUser returns the permission set granted to a user via their role.
	PermissionsForUser(ctx context.Context, userID string) ([]string, error)
	// AssignUserRole sets a user's role_id and keeps the legacy role column in sync.
	AssignUserRole(ctx context.Context, userID, roleID, legacyRole string) error
	// CountUsersWithRole counts active users currently assigned to a role.
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
		SELECT rp.permission
		FROM users u
		JOIN role_permissions rp ON rp.role_id = u.role_id
		WHERE u.id = ?`, userID)
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

func (r *sqliteRoleRepo) AssignUserRole(ctx context.Context, userID, roleID, legacyRole string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET role_id = ?, role = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		roleID, legacyRole, time.Now().Unix(), userID)
	if err != nil {
		return fmt.Errorf("assign user role: %w", err)
	}
	return nil
}

func (r *sqliteRoleRepo) CountUsersWithRole(ctx context.Context, roleID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE role_id = ? AND deleted_at IS NULL`, roleID).Scan(&n)
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
