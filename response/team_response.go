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

type InviteTokenResponse struct {
	TeamID    string    `json:"team_id"    example:"550e8400-e29b-41d4-a716-446655440000"`
	Token     string    `json:"token"      example:"a1b2c3d4e5f6789012345678901234567890123456789012345678901234"`
	InviteURL string    `json:"invite_url" example:"/teams/join?token=a1b2c3..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-01-08T00:00:00Z"`
}

type JoinTeamResponse struct {
	TeamID   string    `json:"team_id"   example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID   string    `json:"user_id"   example:"550e8400-e29b-41d4-a716-446655440001"`
	JoinedAt time.Time `json:"joined_at" example:"2024-01-01T00:00:00Z"`
}

type TeamMemberSkill struct {
	SkillName       string  `json:"skill_name"        example:"Go"`
	ExperienceYears *string `json:"experience_years"  example:"1.5"`
	IsLearningGoal  bool    `json:"is_learning_goal"  example:"false"`
}

type TeamMemberResponse struct {
	ID             string            `json:"id"              example:"550e8400-e29b-41d4-a716-446655440000"`
	Name           string            `json:"name"            example:"山田太郎"`
	AvatarURL      *string           `json:"avatar_url"      example:"https://example.com/avatar.png"`
	SkillLevel     string            `json:"skill_level"     example:"beginner"`
	TeamRole       string            `json:"team_role"       example:"member"`
	FunctionalRole *string           `json:"functional_role" example:"backend"`
	JoinedAt       time.Time         `json:"joined_at"       example:"2024-01-01T00:00:00Z"`
	Skills         []TeamMemberSkill `json:"skills"`
}
