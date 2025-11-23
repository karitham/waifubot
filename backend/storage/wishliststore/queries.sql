-- name: AddCharacterToWishlist :exec
INSERT INTO character_wishlist (user_id, character_id)
VALUES ($1, $2)
ON CONFLICT (user_id, character_id) DO NOTHING;

-- name: AddMultipleCharactersToWishlist :exec
INSERT INTO character_wishlist (user_id, character_id)
SELECT $1, unnest($2::bigint[])
ON CONFLICT (user_id, character_id) DO NOTHING;

-- name: RemoveCharacterFromWishlist :exec
DELETE FROM character_wishlist
WHERE user_id = $1 AND character_id = $2;

-- name: RemoveMultipleCharactersFromWishlist :exec
DELETE FROM character_wishlist
WHERE user_id = $1 AND character_id = ANY($2::bigint[]);

-- name: GetUserCharacterWishlist :many
SELECT
    c.id,
    c.name,
    c.image,
    cw.created_at as date
FROM character_wishlist cw
JOIN characters c ON cw.character_id = c.id
WHERE cw.user_id = $1
ORDER BY cw.created_at DESC;

-- name: GetWishlistHolders :many
WITH user_counts AS (
    SELECT col.user_id, COUNT(*) as match_count
    FROM collection col
    LEFT JOIN guild_members gm ON gm.user_id = col.user_id AND gm.guild_id = $3
    WHERE col.character_id = ANY($1::bigint[])
    AND col.user_id != $2
    AND ($3 = 0 OR gm.guild_id IS NOT NULL)
    GROUP BY col.user_id
    ORDER BY match_count DESC, col.user_id ASC
    LIMIT 20
)
SELECT uc.user_id, c.id as character_id, c.name as character_name, c.image as character_image
FROM user_counts uc
JOIN collection col ON col.user_id = uc.user_id AND col.character_id = ANY($1::bigint[])
JOIN characters c ON col.character_id = c.id
ORDER BY uc.match_count DESC, uc.user_id ASC, c.id ASC;

-- name: GetWantedCharacters :many
WITH user_counts AS (
    SELECT cw.user_id, COUNT(*) as match_count
    FROM character_wishlist cw
    JOIN collection col ON col.character_id = cw.character_id AND col.user_id = $1
    LEFT JOIN guild_members gm ON gm.user_id = cw.user_id AND gm.guild_id = $2
    WHERE cw.user_id != $1
    AND ($2 = 0 OR gm.guild_id IS NOT NULL)
    GROUP BY cw.user_id
    ORDER BY match_count DESC, cw.user_id ASC
    LIMIT 20
)
SELECT uc.user_id, c.id as character_id, c.name as character_name, c.image as character_image
FROM user_counts uc
JOIN character_wishlist cw ON cw.user_id = uc.user_id
JOIN collection col ON col.character_id = cw.character_id AND col.user_id = $1
JOIN characters c ON cw.character_id = c.id
ORDER BY uc.match_count DESC, uc.user_id ASC, c.id ASC;

-- name: CompareWithUser :many
SELECT
    'has' as type,
    c.id,
    c.name,
    c.image,
    cw.created_at as date
FROM character_wishlist cw
JOIN characters c ON cw.character_id = c.id
JOIN collection col ON col.character_id = c.id AND col.user_id = $2
WHERE cw.user_id = $1
UNION ALL
SELECT
    'wants' as type,
    c.id,
    c.name,
    c.image,
    cw.created_at as date
FROM character_wishlist cw
JOIN characters c ON cw.character_id = c.id
WHERE cw.user_id = $2
AND cw.character_id IN (
    SELECT col.character_id
    FROM collection col
    WHERE col.user_id = $1
);