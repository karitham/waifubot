-- name: List :many
SELECT 
    c.id,
    c.name,
    c.image,
    col.source,
    col.acquired_at as date
FROM collection col
JOIN characters c ON col.character_id = c.id
WHERE col.user_id = $1
ORDER BY col.acquired_at DESC;

-- name: ListIDs :many
SELECT 
    col.character_id as id
FROM collection col
WHERE col.user_id = $1;

-- name: SearchCharacters :many
SELECT 
    c.id,
    c.name,
    c.image,
    col.source,
    col.acquired_at as date
FROM collection col
JOIN characters c ON col.character_id = c.id
WHERE col.user_id = sqlc.arg(user_id)
AND (
    c.id::VARCHAR LIKE sqlc.arg(term)::VARCHAR || '%'
    OR c.name ILIKE '%' || sqlc.arg(term) || '%'
)
ORDER BY col.acquired_at DESC
LIMIT sqlc.arg(lim)
OFFSET sqlc.arg(off);

-- name: Get :one
SELECT 
    c.id,
    c.name,
    c.image,
    col.source,
    col.acquired_at as date
FROM collection col
JOIN characters c ON col.character_id = c.id
WHERE c.id = $1
AND col.user_id = $2;

-- name: GetByID :one
SELECT 
    id,
    name,
    image
FROM characters
WHERE id = $1
LIMIT 1;

-- name: Insert :one
INSERT INTO collection (user_id, character_id, source, acquired_at)
VALUES ($1, $2, $3, $4)
RETURNING user_id, character_id, source, acquired_at;

-- name: Give :one
UPDATE collection col
SET 
    user_id = $1,
    source = 'TRADE'
WHERE col.character_id = $2
AND col.user_id = $3
RETURNING user_id, character_id, source, acquired_at;

-- name: Count :one
SELECT 
    COUNT(col.character_id)
FROM collection col
WHERE col.user_id = $1;

-- name: Delete :one
DELETE FROM collection col
WHERE col.user_id = $1
AND col.character_id = $2
RETURNING user_id, character_id, source, acquired_at;

-- name: UpdateImageName :one
UPDATE characters c
SET 
    image = $1,
    name = $2
WHERE c.id = $3
RETURNING id, name, image;

-- name: SearchGlobalCharacters :many
SELECT DISTINCT 
    c.id,
    c.name,
    c.image
FROM characters c
WHERE 
    c.id::VARCHAR LIKE sqlc.arg(term)::VARCHAR || '%'
    OR c.name ILIKE '%' || sqlc.arg(term) || '%'
ORDER BY c.id
LIMIT sqlc.arg(lim);

-- name: UsersOwningCharFiltered :many
SELECT DISTINCT 
    col.user_id
FROM collection col
WHERE col.character_id = $1
AND col.user_id = ANY (sqlc.arg(user_ids)::bigint[]);

-- name: UpsertCharacter :one
INSERT INTO characters (id, name, image)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    image = EXCLUDED.image
RETURNING id, name, image;