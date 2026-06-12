-- +goose Up
-- prefs: free-form JSON blob of client preferences (theme mode, effects,
-- blur levels…) so they follow the account across devices. The server
-- only validates it is JSON and caps its size; the shape belongs to the
-- frontend.
ALTER TABLE users ADD COLUMN prefs TEXT NOT NULL DEFAULT '{}';

-- +goose Down
SELECT 1;
