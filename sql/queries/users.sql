-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET skill_level = $2, bio = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListUserSkills :many
SELECT * FROM user_skills
WHERE user_id = $1
ORDER BY created_at;

-- name: DeleteUserSkills :exec
DELETE FROM user_skills
WHERE user_id = $1;

-- name: CreateUserSkill :one
INSERT INTO user_skills (user_id, skill_name, experience_years, is_learning_goal)
VALUES ($1, $2, $3, $4)
RETURNING *;
