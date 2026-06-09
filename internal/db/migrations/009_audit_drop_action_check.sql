-- +goose Up
-- Remove the rigid CHECK constraint on audit_log.action so new action types
-- (register_blocked, sso_login, settings_changed, invitation_*, etc.) can be
-- logged without a migration each time. The application owns the action values.
ALTER TABLE audit_log RENAME TO audit_log_old;

CREATE TABLE audit_log (
    id         TEXT NOT NULL PRIMARY KEY,
    user_id    TEXT REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT NOT NULL,
    ip         TEXT,
    user_agent TEXT,
    metadata   TEXT,
    created_at INTEGER NOT NULL
);

INSERT INTO audit_log (id, user_id, action, ip, user_agent, metadata, created_at)
SELECT id, user_id, action, ip, user_agent, metadata, created_at FROM audit_log_old;

DROP TABLE audit_log_old;

CREATE INDEX idx_audit_log_user_id    ON audit_log(user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_log_action     ON audit_log(action);

-- +goose Down
-- Irreversible (cannot restore the original CHECK without data loss risk); no-op.
SELECT 1;
