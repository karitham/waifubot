-- name: List :many
SELECT
  c.id,
  c.name,
  c.image,
  c.media_title,
  c.favorites,
  col.source,
  col.acquired_at AS date
FROM
  collection col
  JOIN characters c ON col.character_id = c.id
WHERE
  col.user_id = $1
ORDER BY
  col.acquired_at DESC;

-- name: ListIDs :many
SELECT
  col.character_id AS id
FROM
  collection col
WHERE
  col.user_id = $1;

-- name: SearchCharacters :many
SELECT
  c.id,
  c.name,
  c.image,
  c.media_title,
  c.favorites,
  col.source,
  col.acquired_at AS date
FROM
  collection col
  JOIN characters c ON col.character_id = c.id
WHERE
  col.user_id = sqlc.arg (user_id)
  AND (
    c.id::VARCHAR LIKE sqlc.arg (term)::VARCHAR || '%'
    OR c.name ILIKE '%' || sqlc.arg (term) || '%'
  )
ORDER BY
  col.acquired_at DESC
LIMIT
  sqlc.arg (lim)
OFFSET
  sqlc.arg (off);

-- name: Get :one
SELECT
  c.id,
  c.name,
  c.image,
  c.media_title,
  c.favorites,
  col.source,
  col.acquired_at AS date
FROM
  collection col
  JOIN characters c ON col.character_id = c.id
WHERE
  c.id = $1
  AND col.user_id = $2;

-- name: GetByID :one
SELECT
  id,
  name,
  image,
  media_title,
  favorites,
  updated_at
FROM
  characters
WHERE
  id = $1
LIMIT
  1;

-- name: Insert :one
INSERT INTO
  collection (user_id, character_id, source, acquired_at)
VALUES
  ($1, $2, $3, $4)
RETURNING
  user_id,
  character_id,
  source,
  acquired_at;

-- name: Give :one
UPDATE collection col
SET
  user_id = $1,
  source = 'TRADE',
  acquired_at = NOW()
WHERE
  col.character_id = $2
  AND col.user_id = $3
RETURNING
  user_id,
  character_id,
  source,
  acquired_at;

-- name: Count :one
SELECT
  COUNT(col.character_id)
FROM
  collection col
WHERE
  col.user_id = $1;

-- name: Delete :one
DELETE FROM collection col
WHERE
  col.user_id = $1
  AND col.character_id = $2
RETURNING
  user_id,
  character_id,
  source,
  acquired_at;

-- name: UpdateImageName :one
UPDATE characters c
SET
  image = $1,
  name = $2
WHERE
  c.id = $3
RETURNING
  *;

-- name: SearchGlobalCharacters :many
SELECT DISTINCT
  *
FROM
  characters c
WHERE
  c.id::VARCHAR LIKE sqlc.arg (term)::VARCHAR || '%'
  OR c.name ILIKE '%' || sqlc.arg (term) || '%'
ORDER BY
  c.id
LIMIT
  sqlc.arg (lim);

-- name: UsersOwningCharFiltered :many
SELECT DISTINCT
  col.user_id
FROM
  collection col
WHERE
  col.character_id = $1
  AND col.user_id = ANY (sqlc.arg (user_ids)::BIGINT[]);

-- name: UpsertCharacter :one
INSERT INTO
  characters (id, name, image, favorites)
VALUES
  ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET
  name = excluded.name,
  image = excluded.image,
  favorites = excluded.favorites
RETURNING
  *;

-- name: GetStaleCharacters :many
SELECT id, name, image, media_title, favorites, updated_at
FROM characters
WHERE (updated_at, id) > (sqlc.arg(updated_at), sqlc.arg(cursor_id)::bigint)
  AND updated_at < NOW() - interval '24 hours'
ORDER BY updated_at, id
LIMIT sqlc.arg(lim);

-- name: UpdateCharacterSync :one
UPDATE characters
SET name = $1, image = $2, media_title = $3, favorites = $4, updated_at = NOW()
WHERE id = $5
RETURNING id, name, image, media_title, favorites, updated_at;
