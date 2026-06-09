-- +goose Up
CREATE TABLE users (
    id                TEXT    NOT NULL PRIMARY KEY,
    username          TEXT    NOT NULL UNIQUE,
    email             TEXT    NOT NULL UNIQUE COLLATE NOCASE,
    password          TEXT    NOT NULL,
    role              TEXT    NOT NULL DEFAULT 'user'
                              CHECK(role IN ('user', 'admin')),
    is_active         INTEGER NOT NULL DEFAULT 1
                              CHECK(is_active IN (0, 1)),
    wallpaper_limit   INTEGER,
    search_engine     TEXT    NOT NULL DEFAULT 'duckduckgo'
                              CHECK(search_engine IN (
                                  'duckduckgo', 'google', 'brave',
                                  'bing', 'kagi', 'custom'
                              )),
    search_engine_url TEXT,
    created_at        INTEGER NOT NULL,
    updated_at        INTEGER NOT NULL,
    deleted_at        INTEGER
);

CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_active   ON users(is_active) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE users;
