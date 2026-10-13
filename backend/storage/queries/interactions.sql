-- name: IncrementInteraction :exec
INSERT INTO
  channel_interactions (channel_id, interaction_count)
VALUES
  ($1, 1)
ON CONFLICT (channel_id) DO UPDATE
SET
  interaction_count = channel_interactions.interaction_count + 1;

-- name: GetInteraction :one
SELECT
  interaction_count
FROM
  channel_interactions
WHERE
  channel_id = $1;

-- name: ResetInteraction :exec
DELETE FROM channel_interactions
WHERE
  channel_id = $1;
