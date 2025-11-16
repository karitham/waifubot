-- name: GetGuildMembers :many
SELECT
  user_id
FROM
  guild_members
WHERE
  guild_id = $1;

-- name: UsersOwningCharInGuild :many
SELECT DISTINCT
  c.user_id
FROM
  characters c
  JOIN guild_members gm ON c.user_id = gm.user_id
WHERE
  c.id = $1
  AND gm.guild_id = $2;

-- name: IsGuildIndexed :one
SELECT
  status,
  updated_at
FROM
  guild_indexing_jobs
WHERE
  guild_id = $1;

-- name: GetIndexingStatus :one
SELECT
  status,
  updated_at
FROM
  guild_indexing_jobs
WHERE
  guild_id = $1;

-- name: StartIndexingJob :exec
INSERT INTO
  guild_indexing_jobs (guild_id, status, updated_at)
VALUES
  ($1, 'in_progress', NOW())
ON CONFLICT (guild_id) DO UPDATE
SET
  status = 'in_progress',
  updated_at = NOW();

-- name: CompleteIndexingJob :exec
UPDATE guild_indexing_jobs
SET
  status = 'completed',
  updated_at = NOW()
WHERE
  guild_id = $1;

-- name: DeleteGuildMembers :exec
DELETE FROM guild_members
WHERE
  guild_id = $1;

-- name: DeleteGuildMembersNotIn :exec
DELETE FROM guild_members
WHERE guild_id = $1 AND user_id NOT IN (SELECT unnest($2::bigint[]));

-- name: UpsertGuildMembers :exec
INSERT INTO guild_members (guild_id, user_id, indexed_at)
SELECT $1, unnest($2::bigint[]), $3
ON CONFLICT (guild_id, user_id) DO UPDATE SET indexed_at = EXCLUDED.indexed_at;
