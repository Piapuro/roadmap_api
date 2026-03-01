package requests

type CreateTeamRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateTeamRequest struct {
	Name string `json:"name" validate:"required"`
}
