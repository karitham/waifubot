-- name: AddCharactersToWishlist :exec
INSERT INTO
  character_wishlist (user_id, character_id)
SELECT
  $1,
  UNNEST($2::BIGINT[])
ON CONFLICT (user_id, character_id) DO NOTHING;

-- name: RemoveCharactersFromWishlist :exec
DELETE FROM character_wishlist
WHERE
  user_id = $1
  AND character_id = ANY ($2::BIGINT[]);

-- name: GetUserCharacterWishlist :many
SELECT
  c.id,
  c.name,
  c.image,
  c.favorites,
  cw.created_at AS date
FROM
  character_wishlist cw
  JOIN characters c ON cw.character_id = c.id
WHERE
  cw.user_id = $1
ORDER BY
  cw.created_at DESC;

-- name: GetWishlistHolders :many
WITH
  user_counts AS (
    SELECT
      col.user_id,
      COUNT(*) AS match_count
    FROM
      collection col
      LEFT JOIN guild_members gm ON gm.user_id = col.user_id
      AND gm.guild_id = $3
    WHERE
      col.character_id = ANY ($1::BIGINT[])
      AND col.user_id != $2
      AND (
        $3 = 0
        OR gm.guild_id IS NOT NULL
      )
    GROUP BY
      col.user_id
    ORDER BY
      match_count DESC,
      col.user_id ASC
    LIMIT
      20
  )
SELECT
  uc.user_id,
  c.id AS character_id,
  c.name AS character_name,
  c.image AS character_image
FROM
  user_counts uc
  JOIN collection col ON col.user_id = uc.user_id
  AND col.character_id = ANY ($1::BIGINT[])
  JOIN characters c ON col.character_id = c.id
ORDER BY
  uc.match_count DESC,
  uc.user_id ASC,
  c.id ASC;

-- name: GetWantedCharacters :many
WITH
  user_counts AS (
    SELECT
      cw.user_id,
      COUNT(*) AS match_count
    FROM
      character_wishlist cw
      JOIN collection col ON col.character_id = cw.character_id
      AND col.user_id = $1
      LEFT JOIN guild_members gm ON gm.user_id = cw.user_id
      AND gm.guild_id = $2
    WHERE
      cw.user_id != $1
      AND (
        $2 = 0
        OR gm.guild_id IS NOT NULL
      )
    GROUP BY
      cw.user_id
    ORDER BY
      match_count DESC,
      cw.user_id ASC
    LIMIT
      20
  )
SELECT
  uc.user_id,
  c.id AS character_id,
  c.name AS character_name,
  c.image AS character_image
FROM
  user_counts uc
  JOIN character_wishlist cw ON cw.user_id = uc.user_id
  JOIN collection col ON col.character_id = cw.character_id
  AND col.user_id = $1
  JOIN characters c ON cw.character_id = c.id
ORDER BY
  uc.match_count DESC,
  uc.user_id ASC,
  c.id ASC;

-- name: RemoveAllFromWishlist :exec
DELETE FROM character_wishlist
WHERE
  user_id = $1;

-- name: CompareWithUser :many
SELECT
  'has' AS type,
  c.id,
  c.name,
  c.image,
  cw.created_at AS date
FROM
  character_wishlist cw
  JOIN characters c ON cw.character_id = c.id
  JOIN collection col ON col.character_id = c.id
  AND col.user_id = $2
WHERE
  cw.user_id = $1
UNION ALL
SELECT
  'wants' AS type,
  c.id,
  c.name,
  c.image,
  cw.created_at AS date
FROM
  character_wishlist cw
  JOIN characters c ON cw.character_id = c.id
WHERE
  cw.user_id = $2
  AND cw.character_id IN (
    SELECT
      col.character_id
    FROM
      collection col
    WHERE
      col.user_id = $1
  );

-- name: GetUsersWantingCharacter :many
SELECT
  cw.user_id
FROM
  character_wishlist cw
  LEFT JOIN guild_members gm ON gm.user_id = cw.user_id
  AND gm.guild_id = $2
WHERE
  cw.character_id = $1
  AND cw.user_id != $3
  AND (
    $2 = 0
    OR gm.guild_id IS NOT NULL
  )
ORDER BY
  RANDOM()
LIMIT
  10;
