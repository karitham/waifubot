-- name: CreateToken :exec
INSERT INTO auth_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetTokenWithUser :one
SELECT
    t.token,
    t.user_id,
    t.scopes,
    t.created_at,
    t.expires_at,
    u.id,
    u.user_id,
    u.quote,
    u.date,
    u.favorite,
    u.tokens,
    u.anilist_url,
    u.discord_username,
    u.discord_avatar,
    u.last_updated
FROM auth_tokens t
JOIN users u ON u.user_id = t.user_id
WHERE t.token = $1
  AND t.expires_at > NOW();

-- name: DeleteToken :exec
DELETE FROM auth_tokens WHERE token = $1;
