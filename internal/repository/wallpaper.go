package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type WallpaperRepository interface {
	Create(ctx context.Context, wallpaper *model.Wallpaper) error
	GetByID(ctx context.Context, id, userID string) (*model.Wallpaper, error)
	ListByUser(ctx context.Context, userID string) ([]*model.Wallpaper, error)
	Delete(ctx context.Context, id, userID string) error
	UpdateSort(ctx context.Context, userID string, ids []string) error
	SetPinned(ctx context.Context, id, userID string, pinned bool) error
	CountByUser(ctx context.Context, userID string) (int, error)
}

type sqliteWallpaperRepo struct {
	db *sql.DB
}

func newSQLiteWallpaperRepo(db *sql.DB) WallpaperRepository {
	return &sqliteWallpaperRepo{db: db}
}

func (r *sqliteWallpaperRepo) Create(ctx context.Context, w *model.Wallpaper) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wallpapers (id, user_id, filename, is_pinned, sort, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		w.ID, w.UserID, w.Filename, boolToInt(w.IsPinned), w.Sort, w.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("wallpaper create: %w", err)
	}
	return nil
}

func (r *sqliteWallpaperRepo) GetByID(ctx context.Context, id, userID string) (*model.Wallpaper, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, filename, is_pinned, sort, created_at
		FROM wallpapers WHERE id = ? AND user_id = ?`, id, userID)
	return scanWallpaper(row)
}

func (r *sqliteWallpaperRepo) ListByUser(ctx context.Context, userID string) ([]*model.Wallpaper, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, filename, is_pinned, sort, created_at
		FROM wallpapers WHERE user_id = ? ORDER BY sort ASC, created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("wallpaper list: %w", err)
	}
	defer rows.Close()

	var wallpapers []*model.Wallpaper
	for rows.Next() {
		w, err := scanWallpaper(rows)
		if err != nil {
			return nil, err
		}
		wallpapers = append(wallpapers, w)
	}
	return wallpapers, rows.Err()
}

func (r *sqliteWallpaperRepo) Delete(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM wallpapers WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *sqliteWallpaperRepo) UpdateSort(ctx context.Context, userID string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE wallpapers SET sort = ? WHERE id = ? AND user_id = ?`, i, id, userID,
		); err != nil {
			return fmt.Errorf("wallpaper update sort: %w", err)
		}
	}
	return tx.Commit()
}

func (r *sqliteWallpaperRepo) SetPinned(ctx context.Context, id, userID string, pinned bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE wallpapers SET is_pinned = ? WHERE id = ? AND user_id = ?`,
		boolToInt(pinned), id, userID,
	)
	return err
}

func (r *sqliteWallpaperRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM wallpapers WHERE user_id = ?`, userID,
	).Scan(&n)
	return n, err
}

func scanWallpaper(s scanner) (*model.Wallpaper, error) {
	var w model.Wallpaper
	var isPinned int
	var createdAt int64

	err := s.Scan(&w.ID, &w.UserID, &w.Filename, &isPinned, &w.Sort, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan wallpaper: %w", err)
	}

	w.IsPinned = isPinned == 1
	w.CreatedAt = time.Unix(createdAt, 0)
	return &w, nil
}
