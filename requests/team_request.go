package requests

type CreateTeamRequest struct {
	Name      string `json:"name" validate:"required" example:"Aチーム"`
	Goal      string `json:"goal" validate:"required" example:"Webアプリ開発"`
	Level     string `json:"level" validate:"required" example:"beginner"`
	StartDate string `json:"start_date" validate:"required" example:"2025-01-01"`
	EndDate   string `json:"end_date" validate:"required" example:"2025-03-31"`
}

type UpdateTeamRequest struct {
	Name string `json:"name" validate:"required" example:"Bチーム"`
}
