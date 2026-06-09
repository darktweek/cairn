-- +goose Up
CREATE TABLE wallpapers (
    id          TEXT    NOT NULL PRIMARY KEY,
    user_id     TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename    TEXT    NOT NULL,
    is_pinned   INTEGER NOT NULL DEFAULT 0
                        CHECK(is_pinned IN (0, 1)),
    sort        INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL,
    UNIQUE(user_id, filename)
);

CREATE INDEX idx_wallpapers_user_id ON wallpapers(user_id);

-- +goose Down
DROP TABLE wallpapers;
