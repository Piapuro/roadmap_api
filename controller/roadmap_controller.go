package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
)

type RoadmapController struct {
	roadmapService *service.RoadmapService
}

func NewRoadmapController(roadmapService *service.RoadmapService) *RoadmapController {
	return &RoadmapController{roadmapService: roadmapService}
}

func (c *RoadmapController) CreateRoadmap(ctx echo.Context) error {
	var req requests.CreateRoadmapRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusCreated, nil)
}

func (c *RoadmapController) GetRoadmaps(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *RoadmapController) GetRoadmap(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *RoadmapController) UpdateRoadmap(ctx echo.Context) error {
	var req requests.UpdateRoadmapRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *RoadmapController) DeleteRoadmap(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusNoContent, nil)
}
