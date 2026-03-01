package requests

type CreateRequirementRequest struct {
	TeamID          string   `json:"team_id"         validate:"required,uuid"`
	ProductType     string   `json:"product_type"    validate:"required"`
	DifficultyLevel int      `json:"difficulty_level" validate:"required,min=1,max=5"`
	FreeText        string   `json:"free_text"`
	SupplementURL   string   `json:"supplement_url"`
	Features        []string `json:"features"`
}

type UpdateRequirementRequest struct {
	ProductType     string   `json:"product_type"`
	DifficultyLevel int      `json:"difficulty_level" validate:"omitempty,min=1,max=5"`
	FreeText        string   `json:"free_text"`
	SupplementURL   string   `json:"supplement_url"`
	Features        []string `json:"features"`
}
