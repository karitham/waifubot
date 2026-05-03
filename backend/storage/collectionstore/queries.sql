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

-- name: RandomCharNotOwned :one
SELECT c.id, c.name, c.image, c.media_title, c.favorites
FROM characters c
WHERE c.is_active = true
  AND NOT EXISTS (
    SELECT 1 FROM collection col
    WHERE col.user_id = $1 AND col.character_id = c.id
  )
ORDER BY -ln(random()) / GREATEST(c.favorites, 1)
LIMIT 1;

-- name: RandomActiveChar :one
SELECT id, name, image, media_title, favorites, is_active, updated_at
FROM characters
WHERE is_active = true
ORDER BY -ln(random()) / GREATEST(favorites, 1)
LIMIT 1;

-- name: MarkCharacterInactive :one
UPDATE characters SET is_active = false WHERE id = $1
RETURNING *;

-- name: GetActiveIDs :many
SELECT id FROM characters WHERE is_active = true;
