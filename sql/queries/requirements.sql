-- name: CreateRequirement :one
INSERT INTO requirements (team_id, product_type, difficulty_level, free_text, supplement_url, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRequirementByID :one
SELECT * FROM requirements
WHERE id = $1;

-- name: ListRequirementsByTeamID :many
SELECT * FROM requirements
WHERE team_id = $1
ORDER BY created_at DESC;

-- name: UpdateRequirement :one
UPDATE requirements
SET product_type     = $2,
    difficulty_level = $3,
    free_text        = $4,
    supplement_url   = $5,
    updated_at       = NOW()
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: LockRequirement :one
UPDATE requirements
SET status     = 'locked',
    updated_at = NOW()
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: CreateRequirementFeature :one
INSERT INTO requirement_features (requirement_id, feature_name)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteRequirementFeatures :exec
DELETE FROM requirement_features
WHERE requirement_id = $1;

-- name: ListRequirementFeatures :many
SELECT * FROM requirement_features
WHERE requirement_id = $1
ORDER BY id;
