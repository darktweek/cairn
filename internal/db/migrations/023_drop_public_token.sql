-- +goose Up
-- Public read-only share links were removed; drop the unused column.
DROP INDEX IF EXISTS idx_collections_public_token;
ALTER TABLE collections DROP COLUMN public_token;

-- +goose Down
ALTER TABLE collections ADD COLUMN public_token TEXT;
CREATE UNIQUE INDEX idx_collections_public_token
    ON collections(public_token) WHERE public_token IS NOT NULL;
