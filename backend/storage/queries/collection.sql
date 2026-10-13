-- name: ListCollection :many
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

-- name: ListCollectionIDs :many
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

-- name: GetCollection :one
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

-- name: GetCharacterByID :one
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

-- name: InsertCollection :one
INSERT INTO
  collection (user_id, character_id, source, acquired_at)
VALUES
  ($1, $2, $3, $4)
RETURNING
  user_id,
  character_id,
  source,
  acquired_at;

-- name: GiveCharacter :one
UPDATE collection col
SET
  user_id = $1,
  source = sqlc.arg(source),
  acquired_at = NOW()
WHERE
  col.character_id = $2
  AND col.user_id = $3
RETURNING
  user_id,
  character_id,
  source,
  acquired_at;

-- name: SetFavoriteIfOwned :execrows
-- Sets or clears the favorite; when setting, the character must be in the
-- user's collection, checked in the same statement.
UPDATE users
SET
  favorite = sqlc.arg (favorite)
WHERE
  users.user_id = sqlc.arg (user_id)
  AND (
    sqlc.arg (favorite)::BIGINT IS NULL
    OR EXISTS (
      SELECT 1
      FROM collection col
      WHERE col.user_id = sqlc.arg (user_id)
        AND col.character_id = sqlc.arg (character_id)
    )
  );

-- name: CountCollection :one
SELECT
  COUNT(col.character_id)
FROM
  collection col
WHERE
  col.user_id = $1;

-- name: DeleteCollection :one
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
  c.is_active = true
  AND (c.id::VARCHAR LIKE sqlc.arg (term)::VARCHAR || '%'
    OR c.name ILIKE '%' || sqlc.arg (term) || '%')
ORDER BY
  c.id
LIMIT
  sqlc.arg (lim);

-- name: UpsertCharacter :one
INSERT INTO
  characters (id, name, image, media_title, favorites)
VALUES
  ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
SET
  name = excluded.name,
  image = excluded.image,
  media_title = excluded.media_title,
  favorites = excluded.favorites
RETURNING
  *;

-- name: RandomCharNotOwned :one
SELECT c.id, c.name, c.image, c.media_title, c.favorites
FROM characters c
WHERE c.is_active = true
  AND NOT EXISTS (
    SELECT 1 FROM collection col
    WHERE col.user_id = sqlc.arg(user_id) AND col.character_id = c.id
  )
ORDER BY -ln(random()) / pow(ln(c.favorites + 10), sqlc.arg(weight_exponent)::double precision)
LIMIT 1;

-- name: RandomActiveChar :one
-- Excludes the default AniList placeholder image (set by AniList when a
-- character has no custom artwork) since drops embed the image publicly.
SELECT id, name, image, media_title, favorites, is_active, updated_at
FROM characters
WHERE is_active = true
  AND image != 'https://s4.anilist.co/file/anilistcdn/character/large/default.jpg'
ORDER BY -ln(random()) / pow(ln(favorites + 10), sqlc.arg(weight_exponent)::double precision)
LIMIT 1;

-- name: MarkCharactersInactive :exec
UPDATE characters SET is_active = false WHERE id = ANY(sqlc.arg(ids)::BIGINT[]) AND is_active = true;

-- name: GetActiveIDs :many
SELECT id FROM characters WHERE is_active = true;

-- name: KnownCharacterIDs :many
SELECT id FROM characters WHERE id = ANY(sqlc.arg(ids)::BIGINT[]);
