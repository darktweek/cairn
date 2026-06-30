-- +goose Up
ALTER TABLE bookmarks ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- SQLite does not support DROP COLUMN before 3.35; left intentionally empty.
