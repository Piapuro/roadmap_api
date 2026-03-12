package response

import "time"

type TeamResponse struct {
	ID        string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string    `json:"name"       example:"Aチーム"`
	Goal      string    `json:"goal"       example:"バックエンド開発力向上"`
	Level     string    `json:"level"      example:"intermediate"`
	StartDate time.Time `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate   time.Time `json:"end_date"   example:"2024-06-30T00:00:00Z"`
	CreatedBy string    `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440001"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}
