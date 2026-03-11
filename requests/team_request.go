package requests

type CreateTeamRequest struct {
	Name string `json:"name" validate:"required" example:"Aチーム"`
}

type UpdateTeamRequest struct {
	Name string `json:"name" validate:"required" example:"Bチーム"`
}
