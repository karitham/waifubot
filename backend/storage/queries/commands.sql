-- name: GetCommandHash :one
SELECT hash FROM command_migrations LIMIT 1;

-- name: SetCommandHash :exec
INSERT INTO command_migrations (hash) VALUES ($1)
ON CONFLICT DO NOTHING;

-- name: UpdateCommandHash :exec
UPDATE command_migrations SET hash = $1, deployed_at = NOW();
