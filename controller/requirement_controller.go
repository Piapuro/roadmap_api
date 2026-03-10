package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
)

type RequirementController struct {
	requirementService *service.RequirementService
}

func NewRequirementController(requirementService *service.RequirementService) *RequirementController {
	return &RequirementController{requirementService: requirementService}
}

func (c *RequirementController) CreateRequirement(ctx echo.Context) error {
	var req requests.CreateRequirementRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusCreated, nil)
}

func (c *RequirementController) GetRequirement(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *RequirementController) UpdateRequirement(ctx echo.Context) error {
	var req requests.UpdateRequirementRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *RequirementController) SubmitRequirement(ctx echo.Context) error {
	// TODO: implement（draft → submitted へのステータス遷移）
	return ctx.JSON(http.StatusOK, nil)
}
