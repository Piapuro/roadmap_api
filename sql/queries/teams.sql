-- name: CreateTeam :one
INSERT INTO teams (name, owner_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1;

-- name: ListTeamsByOwner :many
SELECT * FROM teams
WHERE owner_id = $1;

-- name: UpdateTeam :one
UPDATE teams
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = $1;
