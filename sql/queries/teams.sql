-- name: CreateTeam :one
INSERT INTO teams (name, goal, level, start_date, end_date, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1;

-- name: ListTeamsByCreatedBy :many
SELECT * FROM teams
WHERE created_by = $1;

-- name: UpdateTeam :one
UPDATE teams
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = $1;
