-- +goose Up

-- collection_shares: grants a user a permission on a collection they don't own.
CREATE TABLE collection_shares (
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    perm          TEXT NOT NULL CHECK(perm IN ('view', 'edit', 'manage')),
    created_at    INTEGER NOT NULL,
    PRIMARY KEY (collection_id, user_id)
);

CREATE INDEX idx_collection_shares_user ON collection_shares(user_id);

-- +goose Down
DROP TABLE collection_shares;
