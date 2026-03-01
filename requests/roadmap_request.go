package requests

type CreateRoadmapRequest struct {
	Title  string `json:"title"   validate:"required"`
	TeamID string `json:"team_id"`
}

type UpdateRoadmapRequest struct {
	Title   string      `json:"title"   validate:"required"`
	Content interface{} `json:"content"`
}
