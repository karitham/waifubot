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

-- name: IncTokens :exec
UPDATE users
SET
  tokens = tokens + 1
WHERE
  user_id = $1;

-- name: ConsumeTokens :one
UPDATE users
SET
  tokens = tokens - $1
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
