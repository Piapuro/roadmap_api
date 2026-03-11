package requests

type CreateRoadmapRequest struct {
	Title  string `json:"title"   validate:"required" example:"ECサイト開発ロードマップ"`
	TeamID string `json:"team_id"                     example:"550e8400-e29b-41d4-a716-446655440000"`
}

type UpdateRoadmapRequest struct {
	Title   string      `json:"title"   validate:"required" example:"ECサイト開発ロードマップ v2"`
	Content interface{} `json:"content"`
}
