-- name: getChars :many
SELECT
    *
FROM
    characters
WHERE
    characters.user_id = $1
ORDER BY
    characters.date DESC;

-- name: getCharsID :many
SELECT
    id
FROM
    characters
WHERE
    user_id = $1;

-- name: getCharsWhoseIDStartWith :many
SELECT
    *
FROM
    characters
WHERE
    characters.user_id = @user_id
    AND characters.id::VARCHAR LIKE @like_str::string
ORDER BY
    characters.date DESC
LIMIT
    @lim
OFFSET
    @off;

-- name: getChar :one
SELECT
    *
FROM
    characters
WHERE
    id = $1
    AND characters.user_id = $2;

-- name: insertChar :exec
INSERT INTO
    characters ("id", "user_id", "image", "name", "type")
VALUES
    ($1, $2, $3, $4, $5);

-- name: giveChar :one
UPDATE characters
SET
    "type" = 'TRADE',
    "user_id" = @given
WHERE
    characters.id = @id
    AND characters.user_id = @giver
RETURNING
    *;

-- name: createUser :exec
INSERT INTO
    users (user_id)
VALUES
    ($1);

-- name: getUser :one
SELECT
    *
FROM
    users
WHERE
    user_id = $1;

-- name: getProfileOverview :one
SELECT
    characters.image AS favorite_image,
    characters.name AS favorite_name,
    characters.id AS favorite_id,
    users.date AS user_date,
    users.quote AS user_quote,
    users.user_id AS user_id,
    users.tokens AS user_tokens,
    users.anilist_url AS user_anilist_url,
    (
        SELECT
            COUNT(id)
        FROM
            characters
        WHERE
            characters.user_id = $1
    ) AS count
FROM
    users
    LEFT JOIN characters ON characters.id = users.favorite
WHERE
    users.user_id = $1;

-- name: addDropToken :exec
UPDATE users
SET
    tokens = tokens + 1
WHERE
    user_id = $1;

-- name: consumeDropTokens :one
UPDATE users
SET
    tokens = tokens - $1
WHERE
    user_id = $2
RETURNING
    *;

-- name: deleteChar :one
DELETE FROM characters
WHERE
    user_id = $1
    AND id = $2
RETURNING
    *;

-- name: SetChar :one
UPDATE characters
SET
    "image" = $1,
    "name" = $2
WHERE
    id = $3
RETURNING
    *;

-- name: getProfile :many
SELECT
    users.user_id,
    users.quote,
    users.date AS roll_date,
    users.favorite,
    users.tokens,
    users.anilist_url,
    characters.id,
    characters.image,
    characters.name,
    characters.date,
    characters.type
FROM
    users
    INNER JOIN characters ON characters.user_id = users.user_id
WHERE
    users.user_id = $1;

-- name: getUserByAnilist :one
SELECT
    users.user_id,
    users.quote,
    users.date AS roll_date,
    users.favorite,
    users.tokens,
    users.anilist_url
FROM
    users
WHERE
    users.anilist_url = $1;
