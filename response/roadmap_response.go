package response

import "time"

type RoadmapResponse struct {
	ID            string     `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000"`
	TeamID        string     `json:"team_id"        example:"550e8400-e29b-41d4-a716-446655440001"`
	RequirementID string     `json:"requirement_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Status        string     `json:"status"         example:"draft"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty" example:"2024-06-01T00:00:00Z"`
	CreatedAt     time.Time  `json:"created_at"     example:"2024-01-01T00:00:00Z"`
	UpdatedAt     time.Time  `json:"updated_at"     example:"2024-01-01T00:00:00Z"`
}
