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

-- name: IssueInviteToken :one
UPDATE teams
SET invite_token = $2, invite_token_expires_at = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetTeamByInviteToken :one
SELECT * FROM teams
WHERE invite_token = $1;

-- name: IsTeamOwner :one
SELECT EXISTS(
    SELECT 1 FROM user_team_roles
    WHERE user_id = $1 AND team_id = $2 AND team_role_id = 2
);

-- name: IsTeamMember :one
SELECT EXISTS(
    SELECT 1 FROM user_team_roles
    WHERE user_id = $1 AND team_id = $2
);

-- name: JoinTeamAsMember :exec
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, 1)
ON CONFLICT DO NOTHING;

-- name: AssignTeamOwner :exec
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetUserTeamRoleID :one
SELECT utr.team_role_id, tr.name AS team_role_name, tr.level AS team_role_level
FROM user_team_roles utr
JOIN team_roles tr ON tr.id = utr.team_role_id
WHERE utr.user_id = $1 AND utr.team_id = $2;

-- name: ListTeamMembers :many
SELECT
    up.id,
    up.name,
    up.avatar_url,
    up.skill_level,
    utr.team_role_id,
    tr.name AS team_role_name,
    utr.functional_role,
    utr.joined_at
FROM user_team_roles utr
JOIN user_profiles up ON up.id = utr.user_id
JOIN team_roles tr ON tr.id = utr.team_role_id
WHERE utr.team_id = $1
ORDER BY utr.joined_at;
