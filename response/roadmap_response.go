package response

import "time"

type RoadmapResponse struct {
	ID            string    `json:"id"`
	TeamID        string    `json:"team_id"`
	RequirementID string    `json:"requirement_id"`
	Status        string    `json:"status"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
