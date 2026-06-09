package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	SoftDelete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.User, int, error)
	Count(ctx context.Context) (int, error)
	IsFirstUser(ctx context.Context) (bool, error)
}

type sqliteUserRepo struct {
	db *sql.DB
}

func newSQLiteUserRepo(db *sql.DB) UserRepository {
	return &sqliteUserRepo{db: db}
}

func (r *sqliteUserRepo) Create(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users
			(id, username, email, password, role, is_active, wallpaper_limit,
			 search_engine, search_engine_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.Email, u.Password, u.Role, boolToInt(u.IsActive),
		u.WallpaperLimit, u.SearchEngine, u.SearchEngineURL,
		u.CreatedAt.Unix(), u.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("user create: %w", err)
	}
	return nil
}

func (r *sqliteUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password, role, is_active, wallpaper_limit,
		       search_engine, search_engine_url, created_at, updated_at, deleted_at
		FROM users WHERE id = ? AND deleted_at IS NULL`, id)
	return scanUser(row)
}

func (r *sqliteUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password, role, is_active, wallpaper_limit,
		       search_engine, search_engine_url, created_at, updated_at, deleted_at
		FROM users WHERE email = ? AND deleted_at IS NULL`, email)
	return scanUser(row)
}

func (r *sqliteUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password, role, is_active, wallpaper_limit,
		       search_engine, search_engine_url, created_at, updated_at, deleted_at
		FROM users WHERE username = ? AND deleted_at IS NULL`, username)
	return scanUser(row)
}

func (r *sqliteUserRepo) Update(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET
			username = ?, email = ?, password = ?, role = ?, is_active = ?,
			wallpaper_limit = ?, search_engine = ?, search_engine_url = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL`,
		u.Username, u.Email, u.Password, u.Role, boolToInt(u.IsActive),
		u.WallpaperLimit, u.SearchEngine, u.SearchEngineURL, u.UpdatedAt.Unix(), u.ID,
	)
	if err != nil {
		return fmt.Errorf("user update: %w", err)
	}
	return nil
}

func (r *sqliteUserRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET deleted_at = ? WHERE id = ?`,
		time.Now().Unix(), id,
	)
	if err != nil {
		return fmt.Errorf("user soft delete: %w", err)
	}
	return nil
}

func (r *sqliteUserRepo) List(ctx context.Context, offset, limit int) ([]*model.User, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("user list count: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, email, password, role, is_active, wallpaper_limit,
		       search_engine, search_engine_url, created_at, updated_at, deleted_at
		FROM users WHERE deleted_at IS NULL
		ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("user list: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *sqliteUserRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	).Scan(&n)
	return n, err
}

func (r *sqliteUserRepo) IsFirstUser(ctx context.Context) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n == 0, err
}

// scanner accepts both *sql.Row and *sql.Rows
type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (*model.User, error) {
	var u model.User
	var isActive int
	var createdAt, updatedAt int64
	var deletedAt sql.NullInt64
	var wallpaperLimit sql.NullInt64
	var searchEngineURL sql.NullString

	err := s.Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&isActive, &wallpaperLimit, &u.SearchEngine, &searchEngineURL,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	u.IsActive = isActive == 1
	u.CreatedAt = time.Unix(createdAt, 0)
	u.UpdatedAt = time.Unix(updatedAt, 0)

	if wallpaperLimit.Valid {
		v := int(wallpaperLimit.Int64)
		u.WallpaperLimit = &v
	}
	if searchEngineURL.Valid {
		u.SearchEngineURL = &searchEngineURL.String
	}
	if deletedAt.Valid {
		t := time.Unix(deletedAt.Int64, 0)
		u.DeletedAt = &t
	}

	return &u, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
