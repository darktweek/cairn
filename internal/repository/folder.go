package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type FolderRepository interface {
	Create(ctx context.Context, f *model.Folder) error
	GetByID(ctx context.Context, id string) (*model.Folder, error)
	Update(ctx context.Context, f *model.Folder) error
	Delete(ctx context.Context, id string) error
	ListByCollection(ctx context.Context, collectionID string) ([]*model.Folder, error)
	// GetOrCreateRoot returns a root folder (parent_id NULL) by name within a
	// collection, creating it if absent. Used by the bookmark importer.
	GetOrCreateRoot(ctx context.Context, collectionID, name string) (*model.Folder, error)
	// GetOrCreate returns a folder by name and optional parent within a collection,
	// creating it if absent. Used by the bookmark importer for nested folders.
	GetOrCreate(ctx context.Context, collectionID string, parentID *string, name string) (*model.Folder, error)
}

type sqliteFolderRepo struct {
	db *sql.DB
}

func newSQLiteFolderRepo(db *sql.DB) FolderRepository {
	return &sqliteFolderRepo{db: db}
}

func (r *sqliteFolderRepo) Create(ctx context.Context, f *model.Folder) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO folders (id, collection_id, parent_id, name, sort, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		f.ID, f.CollectionID, f.ParentID, f.Name, f.Sort, f.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("folder create: %w", err)
	}
	return nil
}

func (r *sqliteFolderRepo) GetByID(ctx context.Context, id string) (*model.Folder, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, collection_id, parent_id, name, sort, created_at
		FROM folders WHERE id = ?`, id)
	return scanFolder(row)
}

func (r *sqliteFolderRepo) Update(ctx context.Context, f *model.Folder) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE folders SET parent_id = ?, name = ?, sort = ? WHERE id = ?`,
		f.ParentID, f.Name, f.Sort, f.ID,
	)
	if err != nil {
		return fmt.Errorf("folder update: %w", err)
	}
	return nil
}

func (r *sqliteFolderRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM folders WHERE id = ?`, id)
	return err
}

func (r *sqliteFolderRepo) ListByCollection(ctx context.Context, collectionID string) ([]*model.Folder, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, collection_id, parent_id, name, sort, created_at
		FROM folders WHERE collection_id = ?
		ORDER BY sort ASC, name COLLATE NOCASE ASC`, collectionID)
	if err != nil {
		return nil, fmt.Errorf("folder list: %w", err)
	}
	defer rows.Close()

	var out []*model.Folder
	for rows.Next() {
		f, err := scanFolder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *sqliteFolderRepo) GetOrCreateRoot(ctx context.Context, collectionID, name string) (*model.Folder, error) {
	return r.GetOrCreate(ctx, collectionID, nil, name)
}

func (r *sqliteFolderRepo) GetOrCreate(ctx context.Context, collectionID string, parentID *string, name string) (*model.Folder, error) {
	var row *sql.Row
	if parentID == nil {
		row = r.db.QueryRowContext(ctx, `
			SELECT id, collection_id, parent_id, name, sort, created_at
			FROM folders WHERE collection_id = ? AND parent_id IS NULL AND name = ? LIMIT 1`,
			collectionID, name)
	} else {
		row = r.db.QueryRowContext(ctx, `
			SELECT id, collection_id, parent_id, name, sort, created_at
			FROM folders WHERE collection_id = ? AND parent_id = ? AND name = ? LIMIT 1`,
			collectionID, *parentID, name)
	}
	f, err := scanFolder(row)
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	f = &model.Folder{
		ID:           newID(),
		CollectionID: collectionID,
		ParentID:     parentID,
		Name:         name,
		CreatedAt:    time.Now(),
	}
	if err := r.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func scanFolder(s scanner) (*model.Folder, error) {
	var f model.Folder
	var parentID sql.NullString
	var createdAt int64

	err := s.Scan(&f.ID, &f.CollectionID, &parentID, &f.Name, &f.Sort, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan folder: %w", err)
	}
	if parentID.Valid {
		f.ParentID = &parentID.String
	}
	f.CreatedAt = time.Unix(createdAt, 0)
	return &f, nil
}
