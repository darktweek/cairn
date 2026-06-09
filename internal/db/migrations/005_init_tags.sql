-- +goose Up
CREATE TABLE tags (
    id      TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    TEXT NOT NULL COLLATE NOCASE,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_tags_user_id ON tags(user_id);

CREATE TABLE bookmark_tags (
    bookmark_id TEXT NOT NULL REFERENCES bookmarks(id) ON DELETE CASCADE,
    tag_id      TEXT NOT NULL REFERENCES tags(id)      ON DELETE CASCADE,
    PRIMARY KEY (bookmark_id, tag_id)
);

-- +goose Down
DROP TABLE bookmark_tags;
DROP TABLE tags;
