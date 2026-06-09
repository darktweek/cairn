package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type TOTPRepository interface {
	Create(ctx context.Context, userID, encryptedSecret string) error
	GetByUserID(ctx context.Context, userID string) (*model.TOTPSecret, error)
	Verify(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string) error
}

type sqliteTOTPRepo struct {
	db *sql.DB
}

func newSQLiteTOTPRepo(db *sql.DB) TOTPRepository {
	return &sqliteTOTPRepo{db: db}
}

func (r *sqliteTOTPRepo) Create(ctx context.Context, userID, encryptedSecret string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO totp_secrets (user_id, secret, is_verified, created_at)
		VALUES (?, ?, 0, ?)
		ON CONFLICT(user_id) DO UPDATE SET secret = excluded.secret, is_verified = 0`,
		userID, encryptedSecret, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("totp create: %w", err)
	}
	return nil
}

func (r *sqliteTOTPRepo) GetByUserID(ctx context.Context, userID string) (*model.TOTPSecret, error) {
	var t model.TOTPSecret
	var isVerified int
	var createdAt int64

	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, secret, is_verified, created_at FROM totp_secrets WHERE user_id = ?`,
		userID,
	).Scan(&t.UserID, &t.Secret, &isVerified, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("totp get: %w", err)
	}

	t.IsVerified = isVerified == 1
	t.CreatedAt = time.Unix(createdAt, 0)
	return &t, nil
}

func (r *sqliteTOTPRepo) Verify(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE totp_secrets SET is_verified = 1 WHERE user_id = ?`, userID)
	return err
}

func (r *sqliteTOTPRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM totp_secrets WHERE user_id = ?`, userID)
	return err
}
