-- +goose Up
CREATE TABLE IF NOT EXISTS pending_registrations (
    id           TEXT PRIMARY KEY,
    username     TEXT NOT NULL,
    email        TEXT NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    totp_secret  TEXT NOT NULL,
    expires_at   INTEGER NOT NULL,
    created_at   INTEGER NOT NULL DEFAULT (unixepoch()),
    completed_at INTEGER
);
CREATE INDEX IF NOT EXISTS idx_pending_reg_token ON pending_registrations(token_hash);
CREATE INDEX IF NOT EXISTS idx_pending_reg_email ON pending_registrations(email);

-- +goose Down
DROP TABLE IF EXISTS pending_registrations;
