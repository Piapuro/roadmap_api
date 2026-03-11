package response

type SkillTagResponse struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type UserSkillResponse struct {
	ID              string   `json:"id"`
	SkillName       string   `json:"skill_name"`
	ExperienceYears *float64 `json:"experience_years"`
	IsLearningGoal  bool     `json:"is_learning_goal"`
}

type MySkillsResponse struct {
	SkillLevel string              `json:"skill_level"`
	Bio        string              `json:"bio"`
	Skills     []UserSkillResponse `json:"skills"`
}
