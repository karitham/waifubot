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

-- name: ListFilterIDPrefix :many
/* sql-formatter-disable */
SELECT
*
FROM
characters
WHERE
user_id = @user_id
AND id::varchar LIKE @id_prefix::varchar
ORDER BY
date DESC
LIMIT
@lim
OFFSET
@off;
/* sql-formatter-enable */
-- name: Get :one
SELECT
    *
FROM
    characters
WHERE
    id = $1
    AND characters.user_id = $2;

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
