-- +goose Up

-- collections: a shareable container of bookmarks. Each user owns exactly one
-- personal collection (is_personal = 1, auto-created); all others are freely
-- created and (from Phase 3) shareable.
CREATE TABLE collections (
    id          TEXT    NOT NULL PRIMARY KEY,
    owner_id    TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    description TEXT,
    color       TEXT,
    icon        TEXT,
    is_personal INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);
CREATE INDEX idx_collections_owner ON collections(owner_id);

-- folders: a hierarchy of folders WITHIN a single collection. parent_id NULL = root.
CREATE TABLE folders (
    id            TEXT    NOT NULL PRIMARY KEY,
    collection_id TEXT    NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    parent_id     TEXT    REFERENCES folders(id) ON DELETE CASCADE,
    name          TEXT    NOT NULL,
    sort          INTEGER NOT NULL DEFAULT 0,
    created_at    INTEGER NOT NULL
);
CREATE INDEX idx_folders_collection ON folders(collection_id);
CREATE INDEX idx_folders_parent     ON folders(parent_id);

-- bookmarks gain a collection + optional folder. collection_id is required at the
-- application layer; it is kept nullable in the schema only to allow the in-place
-- backfill below (SQLite cannot ADD COLUMN NOT NULL without a constant default).
ALTER TABLE bookmarks ADD COLUMN collection_id TEXT REFERENCES collections(id) ON DELETE CASCADE;
ALTER TABLE bookmarks ADD COLUMN folder_id     TEXT REFERENCES folders(id)     ON DELETE SET NULL;
CREATE INDEX idx_bookmarks_collection ON bookmarks(collection_id);
CREATE INDEX idx_bookmarks_folder_id  ON bookmarks(folder_id);

-- One personal collection per existing user.
INSERT INTO collections (id, owner_id, name, description, color, icon, is_personal, created_at, updated_at)
SELECT lower(hex(randomblob(16))), u.id, 'Personal', NULL, NULL, NULL, 1,
       CAST(strftime('%s','now') AS INTEGER), CAST(strftime('%s','now') AS INTEGER)
FROM users u;

-- Each distinct non-empty folder string becomes a root folder in that user's personal collection.
INSERT INTO folders (id, collection_id, parent_id, name, sort, created_at)
SELECT lower(hex(randomblob(16))), c.id, NULL, f.folder, 0, CAST(strftime('%s','now') AS INTEGER)
FROM (SELECT DISTINCT user_id, folder FROM bookmarks WHERE folder IS NOT NULL AND folder <> '') f
JOIN collections c ON c.owner_id = f.user_id AND c.is_personal = 1;

-- Assign every existing bookmark to its owner's personal collection.
UPDATE bookmarks
SET collection_id = (SELECT c.id FROM collections c WHERE c.owner_id = bookmarks.user_id AND c.is_personal = 1)
WHERE collection_id IS NULL;

-- Link bookmarks that had a folder string to the matching migrated folder.
UPDATE bookmarks
SET folder_id = (
    SELECT fo.id FROM folders fo
    JOIN collections c ON c.id = fo.collection_id
    WHERE c.owner_id = bookmarks.user_id AND c.is_personal = 1 AND fo.name = bookmarks.folder
)
WHERE folder IS NOT NULL AND folder <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_bookmarks_folder_id;
DROP INDEX IF EXISTS idx_bookmarks_collection;
ALTER TABLE bookmarks DROP COLUMN folder_id;
ALTER TABLE bookmarks DROP COLUMN collection_id;
DROP TABLE folders;
DROP TABLE collections;
