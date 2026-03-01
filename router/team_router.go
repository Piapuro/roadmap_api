package router

import (
	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/controller"
	"github.com/your-name/roadmap/api/middleware"
)

func RegisterTeamRoutes(e *echo.Echo, c *controller.TeamController, m *middleware.SupabaseAuth) {
	g := e.Group("/teams", m.Verify)
	g.POST("", c.CreateTeam)
	g.GET("", c.GetTeams)
	g.GET("/:id", c.GetTeam)
	g.PUT("/:id", c.UpdateTeam)
	g.DELETE("/:id", c.DeleteTeam)
}
