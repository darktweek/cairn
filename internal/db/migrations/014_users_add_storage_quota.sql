-- +goose Up
-- storage_quota: total media storage allowed for this user in bytes.
-- NULL = use the global CAIRN_STORAGE_QUOTA default.
ALTER TABLE users ADD COLUMN storage_quota INTEGER;

-- +goose Down
SELECT 1;
