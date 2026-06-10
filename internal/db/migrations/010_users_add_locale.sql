-- +goose Up
ALTER TABLE users ADD COLUMN locale TEXT NOT NULL DEFAULT 'fr';

-- +goose Down
-- SQLite does not support DROP COLUMN on older versions; recreate table
CREATE TABLE users_new AS SELECT * FROM users;
ALTER TABLE users_new DROP COLUMN locale;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
