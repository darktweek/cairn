package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type AuditFilter struct {
	UserID *string
	Action *string
	From   *time.Time
	To     *time.Time
}

type AuditRepository interface {
	Log(ctx context.Context, entry *model.AuditEntry) error
	ListByUser(ctx context.Context, userID string, offset, limit int) ([]*model.AuditEntry, int, error)
	List(ctx context.Context, offset, limit int, filter AuditFilter) ([]*model.AuditEntry, int, error)
}

type sqliteAuditRepo struct {
	db *sql.DB
}

func newSQLiteAuditRepo(db *sql.DB) AuditRepository {
	return &sqliteAuditRepo{db: db}
}

func (r *sqliteAuditRepo) Log(ctx context.Context, e *model.AuditEntry) error {
	var metadata string
	if e.Metadata != nil {
		b, err := json.Marshal(e.Metadata)
		if err == nil {
			metadata = string(b)
		}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_log (id, user_id, action, ip, user_agent, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.UserID, e.Action, e.IP, e.UserAgent, metadata, e.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("audit log: %w", err)
	}
	return nil
}

func (r *sqliteAuditRepo) ListByUser(ctx context.Context, userID string, offset, limit int) ([]*model.AuditEntry, int, error) {
	uid := userID
	f := AuditFilter{UserID: &uid}
	return r.List(ctx, offset, limit, f)
}

func (r *sqliteAuditRepo) List(ctx context.Context, offset, limit int, f AuditFilter) ([]*model.AuditEntry, int, error) {
	args := []any{}
	where := []string{}

	if f.UserID != nil {
		where = append(where, "user_id = ?")
		args = append(args, *f.UserID)
	}
	if f.Action != nil {
		where = append(where, "action = ?")
		args = append(args, *f.Action)
	}
	if f.From != nil {
		where = append(where, "created_at >= ?")
		args = append(args, f.From.Unix())
	}
	if f.To != nil {
		where = append(where, "created_at <= ?")
		args = append(args, f.To.Unix())
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM audit_log"+whereClause, countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("audit list count: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, user_id, action, ip, user_agent, metadata, created_at FROM audit_log"+
			whereClause+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("audit list: %w", err)
	}
	defer rows.Close()

	var entries []*model.AuditEntry
	for rows.Next() {
		e, err := scanAuditEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	return entries, total, rows.Err()
}

func scanAuditEntry(s scanner) (*model.AuditEntry, error) {
	var e model.AuditEntry
	var userID sql.NullString
	var metadata sql.NullString
	var createdAt int64

	err := s.Scan(&e.ID, &userID, &e.Action, &e.IP, &e.UserAgent, &metadata, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("scan audit entry: %w", err)
	}

	if userID.Valid {
		e.UserID = &userID.String
	}
	if metadata.Valid && metadata.String != "" {
		if err := json.Unmarshal([]byte(metadata.String), &e.Metadata); err != nil {
			e.Metadata = map[string]any{"raw": metadata.String}
		}
	}
	e.CreatedAt = time.Unix(createdAt, 0)
	return &e, nil
}
