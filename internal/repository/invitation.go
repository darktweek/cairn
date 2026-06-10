package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/darktweek/cairn/internal/model"
)

type InvitationRepository interface {
	Create(ctx context.Context, inv *model.Invitation) error
	GetByTokenHash(ctx context.Context, hash string) (*model.Invitation, error)
	GetByID(ctx context.Context, id string) (*model.Invitation, error)
	List(ctx context.Context) ([]*model.Invitation, error)
	MarkUsed(ctx context.Context, id string, usedAt time.Time) error
	SetTOTPAndUsername(ctx context.Context, id, encryptedTOTPSecret, username string) error
	Delete(ctx context.Context, id string) error
}

type sqliteInvitationRepo struct{ db *sql.DB }

func newSQLiteInvitationRepo(db *sql.DB) InvitationRepository {
	return &sqliteInvitationRepo{db: db}
}

func (r *sqliteInvitationRepo) Create(ctx context.Context, inv *model.Invitation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO invitations(id,email,token_hash,created_by,expires_at,created_at)
		 VALUES(?,?,?,?,?,?)`,
		inv.ID, inv.Email, inv.TokenHash, inv.CreatedBy,
		inv.ExpiresAt.Unix(), inv.CreatedAt.Unix(),
	)
	return err
}

func (r *sqliteInvitationRepo) GetByTokenHash(ctx context.Context, hash string) (*model.Invitation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,email,token_hash,created_by,expires_at,used_at,created_at,totp_secret,username
		 FROM invitations WHERE token_hash=?`, hash)
	return scanInvitation(row)
}

func (r *sqliteInvitationRepo) GetByID(ctx context.Context, id string) (*model.Invitation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,email,token_hash,created_by,expires_at,used_at,created_at,totp_secret,username
		 FROM invitations WHERE id=?`, id)
	return scanInvitation(row)
}

func (r *sqliteInvitationRepo) List(ctx context.Context) ([]*model.Invitation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,email,token_hash,created_by,expires_at,used_at,created_at,totp_secret,username
		 FROM invitations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Invitation
	for rows.Next() {
		inv, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (r *sqliteInvitationRepo) MarkUsed(ctx context.Context, id string, usedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE invitations SET used_at=? WHERE id=?`, usedAt.Unix(), id)
	return err
}

func (r *sqliteInvitationRepo) SetTOTPAndUsername(ctx context.Context, id, encryptedTOTPSecret, username string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE invitations SET totp_secret=?, username=? WHERE id=?`,
		encryptedTOTPSecret, username, id,
	)
	return err
}

func (r *sqliteInvitationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM invitations WHERE id=?`, id)
	return err
}

type invScanner interface {
	Scan(dest ...any) error
}

func scanInvitation(s invScanner) (*model.Invitation, error) {
	var (
		inv         model.Invitation
		expTS       int64
		usedTS      *int64
		creTS       int64
		totpSecret  sql.NullString
		username    sql.NullString
	)
	if err := s.Scan(&inv.ID, &inv.Email, &inv.TokenHash, &inv.CreatedBy,
		&expTS, &usedTS, &creTS, &totpSecret, &username); err != nil {
		return nil, err
	}
	inv.ExpiresAt = time.Unix(expTS, 0)
	inv.CreatedAt = time.Unix(creTS, 0)
	if usedTS != nil {
		t := time.Unix(*usedTS, 0)
		inv.UsedAt = &t
	}
	if totpSecret.Valid {
		inv.TOTPSecret = totpSecret.String
	}
	if username.Valid {
		inv.Username = &username.String
	}
	return &inv, nil
}
