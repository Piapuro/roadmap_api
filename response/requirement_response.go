package response

import "time"

type RequirementResponse struct {
	ID              string    `json:"id"`
	TeamID          string    `json:"team_id"`
	ProductType     string    `json:"product_type"`
	DifficultyLevel int       `json:"difficulty_level"`
	FreeText        string    `json:"free_text"`
	SupplementURL   string    `json:"supplement_url"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
