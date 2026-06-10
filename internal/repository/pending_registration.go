package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type PendingRegistrationRepository interface {
	Create(ctx context.Context, pr *model.PendingRegistration) error
	GetByTokenHash(ctx context.Context, hash string) (*model.PendingRegistration, error)
	MarkCompleted(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
	List(ctx context.Context) ([]*model.PendingRegistration, error)
	Delete(ctx context.Context, id string) error
}

type sqlitePendingRegRepo struct{ db *sql.DB }

func newSQLitePendingRegRepo(db *sql.DB) PendingRegistrationRepository {
	return &sqlitePendingRegRepo{db: db}
}

func (r *sqlitePendingRegRepo) Create(ctx context.Context, pr *model.PendingRegistration) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pending_registrations
			(id, username, email, token_hash, totp_secret, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		pr.ID, pr.Username, pr.Email, pr.TokenHash, pr.TOTPSecret,
		pr.ExpiresAt.Unix(), pr.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("pending_reg create: %w", err)
	}
	return nil
}

func (r *sqlitePendingRegRepo) GetByTokenHash(ctx context.Context, hash string) (*model.PendingRegistration, error) {
	var pr model.PendingRegistration
	var expiresAt, createdAt int64
	var completedAt sql.NullInt64

	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, token_hash, totp_secret, expires_at, created_at, completed_at
		FROM pending_registrations WHERE token_hash = ?`, hash,
	).Scan(&pr.ID, &pr.Username, &pr.Email, &pr.TokenHash, &pr.TOTPSecret,
		&expiresAt, &createdAt, &completedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("pending_reg get: %w", err)
	}
	pr.ExpiresAt = time.Unix(expiresAt, 0)
	pr.CreatedAt = time.Unix(createdAt, 0)
	if completedAt.Valid {
		t := time.Unix(completedAt.Int64, 0)
		pr.CompletedAt = &t
	}
	return &pr, nil
}

func (r *sqlitePendingRegRepo) MarkCompleted(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE pending_registrations SET completed_at = ? WHERE id = ?`,
		time.Now().Unix(), id,
	)
	return err
}

func (r *sqlitePendingRegRepo) List(ctx context.Context) ([]*model.PendingRegistration, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, email, token_hash, totp_secret, expires_at, created_at, completed_at
		FROM pending_registrations
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("pending_reg list: %w", err)
	}
	defer rows.Close()

	var out []*model.PendingRegistration
	for rows.Next() {
		var pr model.PendingRegistration
		var expiresAt, createdAt int64
		var completedAt sql.NullInt64
		if err := rows.Scan(&pr.ID, &pr.Username, &pr.Email, &pr.TokenHash, &pr.TOTPSecret,
			&expiresAt, &createdAt, &completedAt); err != nil {
			return nil, fmt.Errorf("pending_reg scan: %w", err)
		}
		pr.ExpiresAt = time.Unix(expiresAt, 0)
		pr.CreatedAt = time.Unix(createdAt, 0)
		if completedAt.Valid {
			t := time.Unix(completedAt.Int64, 0)
			pr.CompletedAt = &t
		}
		out = append(out, &pr)
	}
	return out, rows.Err()
}

func (r *sqlitePendingRegRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM pending_registrations WHERE id = ?`, id)
	return err
}

func (r *sqlitePendingRegRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM pending_registrations WHERE expires_at < ? AND completed_at IS NULL`,
		time.Now().Unix(),
	)
	return err
}
