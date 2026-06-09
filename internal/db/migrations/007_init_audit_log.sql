-- +goose Up
CREATE TABLE audit_log (
    id         TEXT NOT NULL PRIMARY KEY,
    user_id    TEXT REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT NOT NULL
                    CHECK(action IN (
                        'login', 'logout', 'login_failed',
                        'password_change', 'totp_enabled', 'totp_disabled',
                        'user_created', 'user_deleted', 'user_suspended',
                        'bookmark_import', 'wallpaper_upload', 'wallpaper_delete'
                    )),
    ip         TEXT,
    user_agent TEXT,
    metadata   TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_audit_log_user_id    ON audit_log(user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_log_action     ON audit_log(action);

-- +goose Down
DROP TABLE audit_log;
