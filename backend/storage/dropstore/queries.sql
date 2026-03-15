-- name: UpsertCharacter :exec
INSERT INTO
  characters (id, name, image, media_title)
VALUES
  ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET
  name = excluded.name,
  image = excluded.image,
  media_title = excluded.media_title;

-- name: SetDrop :exec
INSERT INTO
  channel_drops (channel_id, character_id)
VALUES
  ($1, $2)
ON CONFLICT (channel_id) DO UPDATE
SET
  character_id = excluded.character_id;

-- name: GetDrop :one
SELECT
  c.id,
  c.name,
  c.image,
  c.media_title
FROM
  channel_drops cd
  JOIN characters c ON cd.character_id = c.id
WHERE
  cd.channel_id = $1;

-- name: DeleteDrop :exec
DELETE FROM channel_drops
WHERE
  channel_id = $1;

-- name: GetDropForUpdate :one
SELECT
  c.id,
  c.name,
  c.image,
  c.media_title
FROM
  channel_drops cd
  JOIN characters c ON cd.character_id = c.id
WHERE
  cd.channel_id = $1
FOR UPDATE;
