-- +goose Up
-- public_token: when set, the collection is reachable read-only at /s/{token}.
ALTER TABLE collections ADD COLUMN public_token TEXT;
CREATE UNIQUE INDEX idx_collections_public_token
    ON collections(public_token) WHERE public_token IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_collections_public_token;
ALTER TABLE collections DROP COLUMN public_token;
