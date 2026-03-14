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

-- name: ListTeamsByMember :many
SELECT t.id, t.name, t.goal, t.level, t.start_date, t.end_date, t.is_archived,
       t.invite_token, t.invite_token_expires_at, t.created_by, t.created_at, t.updated_at
FROM teams t
JOIN user_team_roles utr ON utr.team_id = t.id
WHERE utr.user_id = $1
ORDER BY t.created_at DESC;

-- name: UpdateTeam :one
UPDATE teams
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = $1;

-- name: AssignTeamOwner :exec
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;
