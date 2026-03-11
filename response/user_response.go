package response

import "time"

type UserResponse struct {
	ID        string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email"      example:"user@example.com"`
	Name      string    `json:"name"       example:"山田太郎"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}
