-- +goose Up
CREATE TABLE totp_secrets (
    user_id     TEXT    NOT NULL PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret      TEXT    NOT NULL,
    is_verified INTEGER NOT NULL DEFAULT 0
                        CHECK(is_verified IN (0, 1)),
    created_at  INTEGER NOT NULL
);

-- +goose Down
DROP TABLE totp_secrets;
