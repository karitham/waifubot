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
  users.anilist_url = $1;
