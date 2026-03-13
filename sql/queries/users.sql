-- name: GetUserByID :one
SELECT id, name, avatar_url, bio, skill_level, created_at, updated_at
FROM user_profiles
WHERE id = $1;

-- name: UpdateUserName :one
UPDATE user_profiles
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, name, avatar_url, bio, skill_level, created_at, updated_at;

-- name: UpdateUserProfile :one
UPDATE user_profiles
SET skill_level = $2, bio = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, name, avatar_url, bio, skill_level, created_at, updated_at;

-- name: ListUserSkills :many
SELECT id, user_id, skill_name, experience_years, is_learning_goal, created_at
FROM user_skills
WHERE user_id = $1
ORDER BY created_at;

-- name: CreateUserSkill :one
INSERT INTO user_skills (user_id, skill_name, experience_years, is_learning_goal)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, skill_name, experience_years, is_learning_goal, created_at;

-- name: DeleteUserSkills :exec
DELETE FROM user_skills
WHERE user_id = $1;

-- name: AssignGlobalRole :exec
INSERT INTO user_global_roles (user_id, global_role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: EnsureUser :exec
INSERT INTO user_profiles (id, name)
VALUES ($1, $2)
ON CONFLICT (id) DO NOTHING;
