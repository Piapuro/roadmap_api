package router

import (
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/Piapuro/roadmap_api/middleware"
)

func RegisterTeamRoutes(e *echo.Echo, c *controller.TeamController, m *middleware.SupabaseAuth) {
	g := e.Group("/teams", m.Verify)
	g.POST("", c.CreateTeam)
	g.GET("", c.GetTeams)
	g.POST("/join", c.JoinTeam)
	g.GET("/:id", c.GetTeam)
	g.PUT("/:id", c.UpdateTeam)
	g.DELETE("/:id", c.DeleteTeam)
	g.POST("/:id/invite", c.IssueInviteToken)
}
