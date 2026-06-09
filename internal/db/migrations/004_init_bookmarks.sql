-- +goose Up
CREATE TABLE bookmarks (
    id          TEXT    NOT NULL PRIMARY KEY,
    user_id     TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url         TEXT    NOT NULL,
    title       TEXT    NOT NULL,
    folder      TEXT,
    sort        INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_folder  ON bookmarks(user_id, folder);
CREATE INDEX idx_bookmarks_sort    ON bookmarks(user_id, sort);

-- +goose Down
DROP TABLE bookmarks;
