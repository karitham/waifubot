-- name: List :many
SELECT
  *
FROM
  characters
WHERE
  characters.user_id = $1
ORDER BY
  characters.date DESC;

-- name: ListIDs :many
SELECT
  id
FROM
  characters
WHERE
  user_id = $1;

-- name: SearchCharacters :many
SELECT
  *
FROM
  characters
WHERE
  user_id = sqlc.arg (user_id)
  AND (
    id::VARCHAR LIKE sqlc.arg (term)::VARCHAR || '%'
    OR name ILIKE '%' || sqlc.arg (term) || '%'
  )
ORDER BY
  date DESC
LIMIT
  sqlc.arg (lim)
OFFSET
  sqlc.arg (off);

-- name: Get :one
SELECT
  *
FROM
  characters
WHERE
  id = $1
  AND characters.user_id = $2;

-- name: GetByID :one
SELECT
    *
FROM
    characters
WHERE
    id = $1
LIMIT
    1;

-- name: Insert :exec
INSERT INTO
  characters ("id", "user_id", "image", "name", "type")
VALUES
  ($1, $2, $3, $4, $5);

-- name: Give :one
UPDATE characters
SET
  "type" = 'TRADE',
  "user_id" = $1
WHERE
  characters.id = $2
  AND characters.user_id = $3
RETURNING
  *;

-- name: Count :one
SELECT
  COUNT(id)
FROM
  characters
WHERE
  user_id = $1;

-- name: Delete :one
DELETE FROM characters
WHERE
  user_id = $1
  AND id = $2
RETURNING
  *;

-- name: UpdateImageName :one
UPDATE characters
SET
  "image" = $1,
  "name" = $2
WHERE
  id = $3
RETURNING
  *;

-- name: SearchGlobalCharacters :many
SELECT DISTINCT
  ON (id) id,
  name,
  image,
  type
FROM
  characters
WHERE
  id::VARCHAR LIKE sqlc.arg (term)::VARCHAR || '%'
  OR name ILIKE '%' || sqlc.arg (term) || '%'
ORDER BY
  id,
  date DESC
LIMIT
  sqlc.arg (lim);
    *;

-- name: UsersOwningCharFiltered :many
SELECT DISTINCT
    user_id
FROM
    characters
WHERE
    id = $1
    AND user_id = ANY (sqlc.arg(user_ids)::bigint[]);
