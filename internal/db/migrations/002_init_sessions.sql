-- +goose Up
CREATE TABLE sessions (
    id              TEXT    NOT NULL PRIMARY KEY,
    user_id         TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT    NOT NULL UNIQUE,
    user_agent      TEXT,
    ip              TEXT,
    expires_at      INTEGER NOT NULL,
    created_at      INTEGER NOT NULL,
    is_bookmarklet  INTEGER NOT NULL DEFAULT 0
                            CHECK(is_bookmarklet IN (0, 1))
);

CREATE INDEX idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- +goose Down
DROP TABLE sessions;
