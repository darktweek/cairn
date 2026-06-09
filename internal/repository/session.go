package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByTokenHash(ctx context.Context, hash string) (*model.Session, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	ListByUserID(ctx context.Context, userID string) ([]*model.Session, error)
}

type sqliteSessionRepo struct {
	db *sql.DB
}

func newSQLiteSessionRepo(db *sql.DB) SessionRepository {
	return &sqliteSessionRepo{db: db}
}

func (r *sqliteSessionRepo) Create(ctx context.Context, s *model.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions
			(id, user_id, token_hash, user_agent, ip, expires_at, created_at, is_bookmarklet)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.UserID, s.TokenHash, s.UserAgent, s.IP,
		s.ExpiresAt.Unix(), s.CreatedAt.Unix(), boolToInt(s.IsBookmarklet),
	)
	if err != nil {
		return fmt.Errorf("session create: %w", err)
	}
	return nil
}

func (r *sqliteSessionRepo) GetByTokenHash(ctx context.Context, hash string) (*model.Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, user_agent, ip, expires_at, created_at, is_bookmarklet
		FROM sessions
		WHERE token_hash = ? AND expires_at > ?`, hash, time.Now().Unix())
	return scanSession(row)
}

func (r *sqliteSessionRepo) DeleteByID(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (r *sqliteSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func (r *sqliteSessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE expires_at <= ?`, time.Now().Unix())
	return err
}

func (r *sqliteSessionRepo) ListByUserID(ctx context.Context, userID string) ([]*model.Session, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, token_hash, user_agent, ip, expires_at, created_at, is_bookmarklet
		FROM sessions WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("session list: %w", err)
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func scanSession(s scanner) (*model.Session, error) {
	var sess model.Session
	var isBookmarklet int
	var expiresAt, createdAt int64

	err := s.Scan(
		&sess.ID, &sess.UserID, &sess.TokenHash, &sess.UserAgent, &sess.IP,
		&expiresAt, &createdAt, &isBookmarklet,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan session: %w", err)
	}

	sess.ExpiresAt = time.Unix(expiresAt, 0)
	sess.CreatedAt = time.Unix(createdAt, 0)
	sess.IsBookmarklet = isBookmarklet == 1
	return &sess, nil
}
