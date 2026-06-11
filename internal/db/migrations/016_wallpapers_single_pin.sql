-- +goose Up
-- Repair: SetPinned used to allow several pinned wallpapers per user.
-- Keep only the most recently created pinned one for each user.
UPDATE wallpapers SET is_pinned = 0
WHERE is_pinned = 1
  AND id NOT IN (
    SELECT id FROM (
      SELECT id, MAX(created_at) FROM wallpapers
      WHERE is_pinned = 1 GROUP BY user_id
    )
  );

-- +goose Down
SELECT 1;
