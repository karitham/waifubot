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

-- name: UpsertFromDiscord :one
INSERT INTO users (user_id, discord_username, discord_avatar, last_updated)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (user_id) DO UPDATE
  SET discord_username = EXCLUDED.discord_username,
      discord_avatar   = EXCLUDED.discord_avatar,
      last_updated     = NOW()
RETURNING *;

-- name: SearchUsersSharedGuilds :many
SELECT DISTINCT ON (u.user_id)
  u.user_id,
  u.discord_username,
  u.discord_avatar,
  u.anilist_url
FROM users u
JOIN guild_members gm ON gm.user_id = u.user_id
WHERE gm.guild_id IN (SELECT guild_id FROM guild_members AS my_gm WHERE my_gm.user_id = $1)
  AND u.user_id != $1
  AND u.discord_username != ''
  AND (
    LOWER(u.discord_username) LIKE LOWER($2) || '%'
    OR LOWER(u.anilist_url) LIKE LOWER($2) || '%'
  )
ORDER BY u.user_id, u.discord_username
LIMIT $3;
