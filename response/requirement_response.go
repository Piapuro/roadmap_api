package response

import "time"

type RequirementResponse struct {
	ID              string    `json:"id"               example:"550e8400-e29b-41d4-a716-446655440000"`
	TeamID          string    `json:"team_id"          example:"550e8400-e29b-41d4-a716-446655440001"`
	ProductType     string    `json:"product_type"     example:"web"`
	DifficultyLevel int       `json:"difficulty_level" example:"2"`
	FreeText        string    `json:"free_text"        example:"ECサイトを作りたい"`
	SupplementURL   string    `json:"supplement_url"   example:"https://example.com/spec"`
	Features        []string  `json:"features"         example:"ログイン機能,商品一覧,カート機能"`
	Status          string    `json:"status"           example:"draft"`
	CreatedAt       time.Time `json:"created_at"       example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time `json:"updated_at"       example:"2024-01-01T00:00:00Z"`
}
