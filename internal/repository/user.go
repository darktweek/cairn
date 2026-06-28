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

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	UsernamesByIDs(ctx context.Context, ids []string) (map[string]string, error)
	GetPrefs(ctx context.Context, userID string) (string, error)
	SetPrefs(ctx context.Context, userID, prefs string) error
	Update(ctx context.Context, user *model.User) error
	UpdateLocale(ctx context.Context, userID, locale string) error
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.User, int, error)
	Count(ctx context.Context) (int, error)
	IsFirstUser(ctx context.Context) (bool, error)
	// Search returns active users whose username matches q (id + username only),
	// for the collection share picker.
	Search(ctx context.Context, q string, limit int) ([]*model.User, error)
}

type sqliteUserRepo struct {
	db *sql.DB
}

func newSQLiteUserRepo(db *sql.DB) UserRepository {
	return &sqliteUserRepo{db: db}
}

// userCols / userFrom centralise the SELECT shape so every read resolves the
// user's role_id and role display name via the roles table in one join.
const userCols = `u.id, u.username, u.email, u.password, u.role, u.is_active, u.wallpaper_limit, u.upload_size_limit,
	u.storage_quota, u.search_engine, u.search_engine_url, u.locale, u.created_at, u.updated_at, u.deleted_at,
	u.role_id, rl.name`

const userFrom = ` FROM users u LEFT JOIN roles rl ON rl.id = u.role_id `

func (r *sqliteUserRepo) Create(ctx context.Context, u *model.User) error {
	locale := u.Locale
	if locale == "" {
		locale = "en"
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users
			(id, username, email, password, role, role_id, is_active, wallpaper_limit, upload_size_limit,
			 storage_quota, search_engine, search_engine_url, locale, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.Email, u.Password, u.Role, nullStr(u.RoleID), boolToInt(u.IsActive),
		u.WallpaperLimit, u.UploadSizeLimit, u.StorageQuota, u.SearchEngine, u.SearchEngineURL, locale,
		u.CreatedAt.Unix(), u.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("user create: %w", err)
	}
	// Seed the user_roles junction so the multi-role layer stays consistent.
	if u.RoleID != "" {
		if _, err := r.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)`, u.ID, u.RoleID,
		); err != nil {
			return fmt.Errorf("user create role link: %w", err)
		}
	}
	return nil
}

func (r *sqliteUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userCols+userFrom+`WHERE u.id = ? AND u.deleted_at IS NULL`, id)
	return scanUser(row)
}

func (r *sqliteUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userCols+userFrom+`WHERE u.email = ? COLLATE NOCASE AND u.deleted_at IS NULL`, email)
	return scanUser(row)
}

// GetPrefs returns the raw JSON preferences blob for a user.
func (r *sqliteUserRepo) GetPrefs(ctx context.Context, userID string) (string, error) {
	var prefs string
	err := r.db.QueryRowContext(ctx,
		`SELECT prefs FROM users WHERE id = ? AND deleted_at IS NULL`, userID).Scan(&prefs)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get prefs: %w", err)
	}
	return prefs, nil
}

// SetPrefs replaces the JSON preferences blob for a user.
func (r *sqliteUserRepo) SetPrefs(ctx context.Context, userID, prefs string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET prefs = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		prefs, time.Now().Unix(), userID)
	if err != nil {
		return fmt.Errorf("set prefs: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// UsernamesByIDs resolves usernames for a set of user IDs in one query —
// used by the audit log so actor names never require listing all users.
// Soft-deleted users are included: their past actions keep a name.
func (r *sqliteUserRepo) UsernamesByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username FROM users WHERE id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, fmt.Errorf("usernames by ids: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("usernames by ids scan: %w", err)
		}
		out[id] = name
	}
	return out, rows.Err()
}

func (r *sqliteUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userCols+userFrom+`WHERE u.username = ? COLLATE NOCASE AND u.deleted_at IS NULL`, username)
	return scanUser(row)
}

func (r *sqliteUserRepo) Update(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET
			username = ?, email = ?, password = ?, role = ?, is_active = ?,
			wallpaper_limit = ?, upload_size_limit = ?, storage_quota = ?,
			search_engine = ?, search_engine_url = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL`,
		u.Username, u.Email, u.Password, u.Role, boolToInt(u.IsActive),
		u.WallpaperLimit, u.UploadSizeLimit, u.StorageQuota, u.SearchEngine, u.SearchEngineURL, u.UpdatedAt.Unix(), u.ID,
	)
	if err != nil {
		return fmt.Errorf("user update: %w", err)
	}
	return nil
}

func (r *sqliteUserRepo) UpdateLocale(ctx context.Context, userID, locale string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET locale = ? WHERE id = ? AND deleted_at IS NULL`,
		locale, userID,
	)
	return err
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

	rows, err := r.db.QueryContext(ctx,
		`SELECT `+userCols+userFrom+`WHERE u.deleted_at IS NULL
		ORDER BY u.created_at DESC LIMIT ? OFFSET ?`, limit, offset)
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

func (r *sqliteUserRepo) HardDelete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("hard delete begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, q := range []string{
		`DELETE FROM totp_secrets    WHERE user_id = ?`,
		`DELETE FROM sessions        WHERE user_id = ?`,
		`DELETE FROM bookmark_tags   WHERE bookmark_id IN (SELECT id FROM bookmarks WHERE user_id = ?)`,
		`DELETE FROM bookmarks       WHERE user_id = ?`,
		`DELETE FROM wallpapers      WHERE user_id = ?`,
		`DELETE FROM audit_log       WHERE user_id = ?`,
		`DELETE FROM invitations     WHERE created_by = ?`,
		`DELETE FROM users           WHERE id = ?`,
	} {
		if _, err := tx.ExecContext(ctx, q, id); err != nil {
			return fmt.Errorf("hard delete: %w", err)
		}
	}
	return tx.Commit()
}

func (r *sqliteUserRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	).Scan(&n)
	return n, err
}

func (r *sqliteUserRepo) Search(ctx context.Context, q string, limit int) ([]*model.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username FROM users
		WHERE deleted_at IS NULL AND is_active = 1 AND username LIKE ? COLLATE NOCASE
		ORDER BY username COLLATE NOCASE ASC LIMIT ?`, "%"+q+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("user search: %w", err)
	}
	defer rows.Close()
	var out []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, err
		}
		out = append(out, &u)
	}
	return out, rows.Err()
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
	var wallpaperLimit, uploadSizeLimit, storageQuota sql.NullInt64
	var searchEngineURL, roleID, roleName sql.NullString

	err := s.Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&isActive, &wallpaperLimit, &uploadSizeLimit, &storageQuota, &u.SearchEngine, &searchEngineURL,
		&u.Locale, &createdAt, &updatedAt, &deletedAt,
		&roleID, &roleName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	u.RoleID = roleID.String
	u.RoleName = roleName.String

	u.IsActive = isActive == 1
	u.CreatedAt = time.Unix(createdAt, 0)
	u.UpdatedAt = time.Unix(updatedAt, 0)

	if wallpaperLimit.Valid {
		v := int(wallpaperLimit.Int64)
		u.WallpaperLimit = &v
	}
	if uploadSizeLimit.Valid {
		u.UploadSizeLimit = &uploadSizeLimit.Int64
	}
	if storageQuota.Valid {
		u.StorageQuota = &storageQuota.Int64
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
