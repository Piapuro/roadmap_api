package response

import "time"

// UserResponse は認証系レスポンス（SignUp / Login）で使用する。
// email は Supabase から直接取得できるため含む。
type UserResponse struct {
	ID        string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email"      example:"user@example.com"`
	Name      string    `json:"name"       example:"山田太郎"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ProfileResponse は GET /users/me で返すプロフィール情報。
// user_profiles テーブルから取得するため email は含まない。
type ProfileResponse struct {
	ID         string    `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Name       string    `json:"name"        example:"山田太郎"`
	AvatarURL  *string   `json:"avatar_url"  example:"https://example.com/avatar.png"`
	Bio        string    `json:"bio"         example:"Goエンジニア"`
	SkillLevel string    `json:"skill_level" example:"beginner"`
	CreatedAt  time.Time `json:"created_at"  example:"2024-01-01T00:00:00Z"`
}
