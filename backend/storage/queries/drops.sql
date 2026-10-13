-- name: SetDrop :exec
INSERT INTO
  channel_drops (channel_id, character_id)
VALUES
  ($1, $2)
ON CONFLICT (channel_id) DO UPDATE
SET
  character_id = excluded.character_id;

-- name: DeleteDrop :exec
DELETE FROM channel_drops
WHERE
  channel_id = $1;

-- name: GetDropForUpdate :one
-- Locks only the drop row; locking the joined character row blocks inserts
-- that reference it.
SELECT
  c.id,
  c.name,
  c.image,
  c.media_title,
  c.favorites
FROM
  channel_drops cd
  JOIN characters c ON cd.character_id = c.id
WHERE
  cd.channel_id = $1
FOR UPDATE OF
  cd;
