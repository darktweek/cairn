-- +goose Up
-- The CHECK(action IN (...)) list from 007 silently rejected every action
-- added since (login_sso, registration_requested, invitation_sent, ...)
-- because audit logging is fire-and-forget. The action vocabulary is owned
-- by the application now — rebuild the table without the constraint.
CREATE TABLE audit_log_new (
    id         TEXT NOT NULL PRIMARY KEY,
    user_id    TEXT REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT NOT NULL,
    ip         TEXT,
    user_agent TEXT,
    metadata   TEXT,
    created_at INTEGER NOT NULL
);
INSERT INTO audit_log_new SELECT id, user_id, action, ip, user_agent, metadata, created_at FROM audit_log;
DROP TABLE audit_log;
ALTER TABLE audit_log_new RENAME TO audit_log;
CREATE INDEX idx_audit_log_user_id    ON audit_log(user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_log_action     ON audit_log(action);

-- +goose Down
SELECT 1;
