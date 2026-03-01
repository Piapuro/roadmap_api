package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/requests"
	"github.com/your-name/roadmap/api/service"
)

type TeamController struct {
	teamService *service.TeamService
}

func NewTeamController(teamService *service.TeamService) *TeamController {
	return &TeamController{teamService: teamService}
}

func (c *TeamController) CreateTeam(ctx echo.Context) error {
	var req requests.CreateTeamRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusCreated, nil)
}

func (c *TeamController) GetTeams(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *TeamController) GetTeam(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *TeamController) UpdateTeam(ctx echo.Context) error {
	var req requests.UpdateTeamRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *TeamController) DeleteTeam(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusNoContent, nil)
}
