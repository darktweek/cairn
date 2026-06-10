-- +goose Up
ALTER TABLE users ADD COLUMN upload_size_limit INTEGER;

-- +goose Down
-- SQLite does not support DROP COLUMN in older versions; left as no-op.
SELECT 1;
