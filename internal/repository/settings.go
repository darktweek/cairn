package repository

import (
	"context"
	"database/sql"
)

type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

type sqliteSettingsRepo struct{ db *sql.DB }

func newSQLiteSettingsRepo(db *sql.DB) SettingsRepository {
	return &sqliteSettingsRepo{db: db}
}

func (r *sqliteSettingsRepo) Get(ctx context.Context, key string) (string, error) {
	var v string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key=?`, key).Scan(&v)
	return v, err
}

func (r *sqliteSettingsRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
		key, value)
	return err
}
