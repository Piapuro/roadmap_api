package router

import (
	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/controller"
	"github.com/your-name/roadmap/api/middleware"
)

func RegisterRoadmapRoutes(e *echo.Echo, c *controller.RoadmapController, m *middleware.SupabaseAuth) {
	g := e.Group("/roadmaps", m.Verify)
	g.POST("", c.CreateRoadmap)
	g.GET("", c.GetRoadmaps)
	g.GET("/:id", c.GetRoadmap)
	g.PUT("/:id", c.UpdateRoadmap)
	g.DELETE("/:id", c.DeleteRoadmap)
}
