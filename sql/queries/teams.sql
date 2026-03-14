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

-- name: IssueInviteToken :one
UPDATE teams
SET invite_token = $2, invite_token_expires_at = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetTeamByInviteToken :one
SELECT * FROM teams
WHERE invite_token = $1;

-- name: IsTeamOwner :one
SELECT EXISTS (
    SELECT 1 FROM user_team_roles utr
    JOIN team_roles tr ON tr.id = utr.team_role_id
    WHERE utr.user_id = $1 AND utr.team_id = $2 AND tr.level >= 20
) AS is_owner;

-- name: IsTeamMember :one
SELECT EXISTS (
    SELECT 1 FROM user_team_roles
    WHERE user_id = $1 AND team_id = $2
) AS is_member;

-- name: JoinTeamAsMember :exec
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, 1)
ON CONFLICT (user_id, team_id) DO NOTHING;

-- name: ListTeamMembers :many
SELECT up.id, up.name, up.avatar_url, up.skill_level,
       tr.name AS team_role_name, utr.joined_at, utr.functional_role
FROM user_team_roles utr
JOIN user_profiles up ON up.id = utr.user_id
JOIN team_roles tr ON tr.id = utr.team_role_id
WHERE utr.team_id = $1
ORDER BY utr.joined_at;

-- name: GetUserTeamRoleID :one
SELECT tr.level AS team_role_level, tr.name AS team_role_name
FROM user_team_roles utr
JOIN team_roles tr ON tr.id = utr.team_role_id
WHERE utr.user_id = $1 AND utr.team_id = $2;
