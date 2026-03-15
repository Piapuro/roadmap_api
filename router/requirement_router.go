package router

import (
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/Piapuro/roadmap_api/middleware"
)

func RegisterRequirementRoutes(e *echo.Echo, c *controller.RequirementController, rc *controller.RoadmapController, m *middleware.SupabaseAuth) {
	g := e.Group("/requirements", m.Verify)
	g.GET("/:id", c.GetRequirement)
	g.PUT("/:id", c.UpdateRequirement)
	g.POST("/:id/submit", c.SubmitRequirement)
	g.POST("/:id/suggest-mvp", rc.SuggestMVP) // AI MVP提案
}
