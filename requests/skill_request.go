package requests

type SkillInput struct {
	SkillName       string   `json:"skill_name"       validate:"required,max=30"`
	ExperienceYears *float64 `json:"experience_years"`
	IsLearningGoal  bool     `json:"is_learning_goal"`
}

type UpsertSkillsRequest struct {
	SkillLevel string       `json:"skill_level" validate:"required,oneof=beginner intermediate advanced"`
	Bio        string       `json:"bio"         validate:"max=200"`
	Skills     []SkillInput `json:"skills"      validate:"required"`
}
