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

type BookmarkFilter struct {
	Folder *string
	TagID  *string
	Search string
	Offset int
	Limit  int
}

type BookmarkRepository interface {
	Create(ctx context.Context, bookmark *model.Bookmark) error
	GetByID(ctx context.Context, id, userID string) (*model.Bookmark, error)
	Update(ctx context.Context, bookmark *model.Bookmark) error
	Delete(ctx context.Context, id, userID string) error
	ListByUser(ctx context.Context, userID string, filter BookmarkFilter) ([]*model.Bookmark, int, error)
	UpdateSort(ctx context.Context, userID string, ids []string) error
	CountByUser(ctx context.Context, userID string) (int, error)
	AddTag(ctx context.Context, bookmarkID, tagID string) error
	RemoveTag(ctx context.Context, bookmarkID, tagID string) error
	SetTags(ctx context.Context, bookmarkID string, tagIDs []string) error
	BulkCreate(ctx context.Context, userID string, bookmarks []*model.Bookmark) error
}

type sqliteBookmarkRepo struct {
	db *sql.DB
}

func newSQLiteBookmarkRepo(db *sql.DB) BookmarkRepository {
	return &sqliteBookmarkRepo{db: db}
}

func (r *sqliteBookmarkRepo) Create(ctx context.Context, b *model.Bookmark) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO bookmarks (id, user_id, url, title, folder, sort, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.UserID, b.URL, b.Title, b.Folder, b.Sort,
		b.CreatedAt.Unix(), b.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("bookmark create: %w", err)
	}
	return nil
}

func (r *sqliteBookmarkRepo) GetByID(ctx context.Context, id, userID string) (*model.Bookmark, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, url, title, folder, sort, created_at, updated_at
		FROM bookmarks WHERE id = ? AND user_id = ?`, id, userID)
	b, err := scanBookmark(row)
	if err != nil {
		return nil, err
	}
	if err := r.loadTags(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *sqliteBookmarkRepo) Update(ctx context.Context, b *model.Bookmark) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE bookmarks SET url = ?, title = ?, folder = ?, sort = ?, updated_at = ?
		WHERE id = ? AND user_id = ?`,
		b.URL, b.Title, b.Folder, b.Sort, b.UpdatedAt.Unix(), b.ID, b.UserID,
	)
	return err
}

func (r *sqliteBookmarkRepo) Delete(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM bookmarks WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *sqliteBookmarkRepo) ListByUser(ctx context.Context, userID string, f BookmarkFilter) ([]*model.Bookmark, int, error) {
	args := []any{userID}
	where := []string{"user_id = ?"}

	if f.Folder != nil {
		where = append(where, "folder = ?")
		args = append(args, *f.Folder)
	}
	if f.Search != "" {
		where = append(where, "(title LIKE ? OR url LIKE ?)")
		term := "%" + f.Search + "%"
		args = append(args, term, term)
	}

	baseQuery := "FROM bookmarks"
	if f.TagID != nil {
		baseQuery = "FROM bookmarks JOIN bookmark_tags bt ON bt.bookmark_id = bookmarks.id"
		where = append(where, "bt.tag_id = ?")
		args = append(args, *f.TagID)
	}

	whereClause := " WHERE " + strings.Join(where, " AND ")

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) "+baseQuery+whereClause, countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("bookmark list count: %w", err)
	}

	args = append(args, f.Limit, f.Offset)
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, url, title, folder, sort, created_at, updated_at "+
			baseQuery+whereClause+
			" ORDER BY sort ASC, created_at DESC LIMIT ? OFFSET ?",
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("bookmark list: %w", err)
	}
	defer rows.Close()

	var bookmarks []*model.Bookmark
	for rows.Next() {
		b, err := scanBookmark(rows)
		if err != nil {
			return nil, 0, err
		}
		bookmarks = append(bookmarks, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for _, b := range bookmarks {
		if err := r.loadTags(ctx, b); err != nil {
			return nil, 0, err
		}
	}

	return bookmarks, total, nil
}

func (r *sqliteBookmarkRepo) UpdateSort(ctx context.Context, userID string, ids []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE bookmarks SET sort = ? WHERE id = ? AND user_id = ?`, i, id, userID,
		); err != nil {
			return fmt.Errorf("bookmark update sort: %w", err)
		}
	}
	return tx.Commit()
}

func (r *sqliteBookmarkRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM bookmarks WHERE user_id = ?`, userID,
	).Scan(&n)
	return n, err
}

func (r *sqliteBookmarkRepo) AddTag(ctx context.Context, bookmarkID, tagID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`,
		bookmarkID, tagID,
	)
	return err
}

func (r *sqliteBookmarkRepo) RemoveTag(ctx context.Context, bookmarkID, tagID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM bookmark_tags WHERE bookmark_id = ? AND tag_id = ?`,
		bookmarkID, tagID,
	)
	return err
}

func (r *sqliteBookmarkRepo) SetTags(ctx context.Context, bookmarkID string, tagIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM bookmark_tags WHERE bookmark_id = ?`, bookmarkID,
	); err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`,
			bookmarkID, tagID,
		); err != nil {
			return fmt.Errorf("set tag: %w", err)
		}
	}
	return tx.Commit()
}

func (r *sqliteBookmarkRepo) BulkCreate(ctx context.Context, userID string, bookmarks []*model.Bookmark) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO bookmarks (id, user_id, url, title, folder, sort, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range bookmarks {
		if _, err := stmt.ExecContext(ctx,
			b.ID, b.UserID, b.URL, b.Title, b.Folder, b.Sort,
			b.CreatedAt.Unix(), b.UpdatedAt.Unix(),
		); err != nil {
			return fmt.Errorf("bulk create bookmark: %w", err)
		}
	}
	return tx.Commit()
}

func (r *sqliteBookmarkRepo) loadTags(ctx context.Context, b *model.Bookmark) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.user_id, t.name
		FROM tags t
		JOIN bookmark_tags bt ON bt.tag_id = t.id
		WHERE bt.bookmark_id = ?`, b.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name); err != nil {
			return err
		}
		b.Tags = append(b.Tags, t)
	}
	return rows.Err()
}

func scanBookmark(s scanner) (*model.Bookmark, error) {
	var b model.Bookmark
	var folder sql.NullString
	var createdAt, updatedAt int64

	err := s.Scan(&b.ID, &b.UserID, &b.URL, &b.Title, &folder, &b.Sort, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan bookmark: %w", err)
	}

	if folder.Valid {
		b.Folder = &folder.String
	}
	b.CreatedAt = time.Unix(createdAt, 0)
	b.UpdatedAt = time.Unix(updatedAt, 0)
	return &b, nil
}
