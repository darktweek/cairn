package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type GroupRepository interface {
	ListAll(ctx context.Context) ([]*model.Group, error)
	GetByID(ctx context.Context, id string) (*model.Group, error)
	Create(ctx context.Context, g *model.Group) error
	Update(ctx context.Context, g *model.Group) error
	Delete(ctx context.Context, id string) error

	ListMembers(ctx context.Context, groupID string) ([]*model.GroupMember, error)
	SetMember(ctx context.Context, groupID, userID, role string) error
	RemoveMember(ctx context.Context, groupID, userID string) error
}

type sqliteGroupRepo struct {
	db *sql.DB
}

func newSQLiteGroupRepo(db *sql.DB) GroupRepository {
	return &sqliteGroupRepo{db: db}
}

func (r *sqliteGroupRepo) ListAll(ctx context.Context) ([]*model.Group, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.owner_id, g.created_at, g.updated_at,
		       (SELECT COUNT(*) FROM group_members gm WHERE gm.group_id = g.id) AS cnt
		FROM groups g ORDER BY g.name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, fmt.Errorf("group list: %w", err)
	}
	defer rows.Close()
	var out []*model.Group
	for rows.Next() {
		g, err := scanGroup(rows, true)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *sqliteGroupRepo) GetByID(ctx context.Context, id string) (*model.Group, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, owner_id, created_at, updated_at FROM groups WHERE id = ?`, id)
	return scanGroup(row, false)
}

func (r *sqliteGroupRepo) Create(ctx context.Context, g *model.Group) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO groups (id, name, owner_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		g.ID, g.Name, g.OwnerID, g.CreatedAt.Unix(), g.UpdatedAt.Unix())
	if err != nil {
		return fmt.Errorf("group create: %w", err)
	}
	return nil
}

func (r *sqliteGroupRepo) Update(ctx context.Context, g *model.Group) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE groups SET name = ?, updated_at = ? WHERE id = ?`, g.Name, time.Now().Unix(), g.ID)
	return err
}

func (r *sqliteGroupRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, id)
	return err
}

func (r *sqliteGroupRepo) ListMembers(ctx context.Context, groupID string) ([]*model.GroupMember, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT gm.user_id, u.username, gm.role
		FROM group_members gm JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY u.username COLLATE NOCASE ASC`, groupID)
	if err != nil {
		return nil, fmt.Errorf("group members: %w", err)
	}
	defer rows.Close()
	var out []*model.GroupMember
	for rows.Next() {
		var m model.GroupMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.Role); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

func (r *sqliteGroupRepo) SetMember(ctx context.Context, groupID, userID, role string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role) VALUES (?, ?, ?)
		ON CONFLICT(group_id, user_id) DO UPDATE SET role = excluded.role`,
		groupID, userID, role)
	if err != nil {
		return fmt.Errorf("set member: %w", err)
	}
	return nil
}

func (r *sqliteGroupRepo) RemoveMember(ctx context.Context, groupID, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, userID)
	return err
}

func scanGroup(s scanner, withCount bool) (*model.Group, error) {
	var g model.Group
	var createdAt, updatedAt int64
	var err error
	if withCount {
		err = s.Scan(&g.ID, &g.Name, &g.OwnerID, &createdAt, &updatedAt, &g.MemberCount)
	} else {
		err = s.Scan(&g.ID, &g.Name, &g.OwnerID, &createdAt, &updatedAt)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan group: %w", err)
	}
	g.CreatedAt = time.Unix(createdAt, 0)
	g.UpdatedAt = time.Unix(updatedAt, 0)
	return &g, nil
}
