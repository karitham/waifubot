-- name: Create :exec
INSERT INTO
  users (user_id)
VALUES
  ($1);

-- name: Get :one
SELECT
  *
FROM
  users
WHERE
  user_id = $1;

-- name: UpdateTokens :one
UPDATE users
SET
  tokens = tokens + $1
WHERE
  user_id = $2
RETURNING
  *;

-- name: GetByAnilist :one
SELECT
  *
FROM
  users
WHERE
  LOWER(users.anilist_url) = LOWER($1);

-- name: GetByDiscordUsername :one
SELECT
  *
FROM
  users
WHERE
  discord_username = $1
  AND discord_username != '';

-- name: UpdateFavorite :exec
UPDATE users
SET
  favorite = $1
WHERE
  user_id = $2;

-- name: UpdateAnilistURL :exec
UPDATE users
SET
  anilist_url = $1
WHERE
  user_id = $2;

-- name: UpdateQuote :exec
UPDATE users
SET
  quote = $1
WHERE
  user_id = $2;

-- name: UpdateDate :exec
UPDATE users
SET
  date = $1
WHERE
  user_id = $2;

-- name: UpdateDiscordInfo :exec
UPDATE users
SET
  discord_username = $1,
  discord_avatar = $2,
  last_updated = $3
WHERE
  user_id = $4;
