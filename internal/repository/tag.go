package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/darktweek/cairn/internal/model"
)

type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) (*model.Tag, error)
	GetOrCreate(ctx context.Context, userID, name string) (*model.Tag, error)
	ListByUser(ctx context.Context, userID string) ([]*model.Tag, error)
	Delete(ctx context.Context, id, userID string) error
}

type sqliteTagRepo struct {
	db *sql.DB
}

func newSQLiteTagRepo(db *sql.DB) TagRepository {
	return &sqliteTagRepo{db: db}
}

func (r *sqliteTagRepo) Create(ctx context.Context, t *model.Tag) (*model.Tag, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tags (id, user_id, name) VALUES (?, ?, ?)`,
		t.ID, t.UserID, t.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("tag create: %w", err)
	}
	return t, nil
}

func (r *sqliteTagRepo) GetOrCreate(ctx context.Context, userID, name string) (*model.Tag, error) {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO tags (id, user_id, name) VALUES (lower(hex(randomblob(16))), ?, ?)`,
		userID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("tag get or create insert: %w", err)
	}

	var t model.Tag
	err = r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name FROM tags WHERE user_id = ? AND name = ?`,
		userID, name,
	).Scan(&t.ID, &t.UserID, &t.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("tag get or create select: %w", err)
	}
	return &t, nil
}

func (r *sqliteTagRepo) ListByUser(ctx context.Context, userID string) ([]*model.Tag, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name FROM tags WHERE user_id = ? ORDER BY name ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("tag list: %w", err)
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, &t)
	}
	return tags, rows.Err()
}

func (r *sqliteTagRepo) Delete(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM tags WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
