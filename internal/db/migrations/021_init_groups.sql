-- +goose Up

-- groups: a named team of users; collections can be shared with a whole group.
CREATE TABLE groups (
    id         TEXT    NOT NULL PRIMARY KEY,
    name       TEXT    NOT NULL,
    owner_id   TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX idx_groups_owner ON groups(owner_id);

CREATE TABLE group_members (
    group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id  TEXT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    role     TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('admin', 'member')),
    PRIMARY KEY (group_id, user_id)
);
CREATE INDEX idx_group_members_user ON group_members(user_id);

-- collection_group_shares: grants a whole group a permission on a collection.
CREATE TABLE collection_group_shares (
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    group_id      TEXT NOT NULL REFERENCES groups(id)      ON DELETE CASCADE,
    perm          TEXT NOT NULL CHECK(perm IN ('view', 'edit', 'manage')),
    created_at    INTEGER NOT NULL,
    PRIMARY KEY (collection_id, group_id)
);
CREATE INDEX idx_collection_group_shares_group ON collection_group_shares(group_id);

-- +goose Down
DROP TABLE collection_group_shares;
DROP TABLE group_members;
DROP TABLE groups;
