package router

import (
	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/controller"
	"github.com/your-name/roadmap/api/middleware"
)

func RegisterRequirementRoutes(e *echo.Echo, c *controller.RequirementController, m *middleware.SupabaseAuth) {
	g := e.Group("/requirements", m.Verify)
	g.POST("", c.CreateRequirement)
	g.GET("/:id", c.GetRequirement)
	g.PUT("/:id", c.UpdateRequirement)
	g.POST("/:id/submit", c.SubmitRequirement)
}
