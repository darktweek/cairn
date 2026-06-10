-- +goose Up
-- Store a temporary TOTP secret so the setup page can show a QR before the user account exists.
ALTER TABLE invitations ADD COLUMN totp_secret TEXT;
ALTER TABLE invitations ADD COLUMN username    TEXT;

-- +goose Down
SELECT 1;
