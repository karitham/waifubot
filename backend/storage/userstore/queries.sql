-- name: Create :exec
INSERT INTO
  users (user_id)
VALUES
  ($1)
ON CONFLICT (user_id) DO NOTHING;

-- name: Get :one
SELECT
  *
FROM
  users
WHERE
  user_id = $1;

-- name: LockUser :one
-- FOR NO KEY UPDATE: serializes writers, does not block foreign key checks.
SELECT
  user_id
FROM
  users
WHERE
  user_id = $1
FOR NO KEY UPDATE;

-- name: UpdateTokens :one
UPDATE users
SET
  tokens = tokens + $1
WHERE
  user_id = $2
RETURNING
  *;

-- name: SpendTokens :one
UPDATE users
SET
  tokens = tokens - $1
WHERE
  user_id = $2
  AND tokens >= $1
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

-- name: ClearFavorite :exec
UPDATE users
SET
  favorite = NULL
WHERE
  user_id = $1
  AND favorite = $2;

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

-- name: RecordRoll :execrows
-- Zero rows affected means the cooldown is still active.
UPDATE users
SET
  date = $1
WHERE
  user_id = $2
  AND date <= sqlc.arg(threshold);

-- name: UpdateDiscordInfo :exec
UPDATE users
SET
  discord_username = $1,
  discord_avatar = $2,
  last_updated = $3
WHERE
  user_id = $4;
