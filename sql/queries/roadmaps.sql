-- name: HasConfirmedRoadmapForTeam :one
-- チームに confirmed 状態のロードマップが存在するか確認する
SELECT EXISTS (
    SELECT 1 FROM roadmaps
    WHERE team_id = $1 AND status = 'confirmed'
) AS exists;
