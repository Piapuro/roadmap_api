package router

import (
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/Piapuro/roadmap_api/middleware"
)

func RegisterUserRoutes(e *echo.Echo, c *controller.UserController, m *middleware.SupabaseAuth) {
	g := e.Group("/users", m.Verify)
	g.GET("/me", c.GetMe)
	g.PUT("/me", c.UpdateMe)
	g.GET("/me/skills", c.GetMySkills)
	g.PUT("/me/skills", c.UpsertMySkills)
}
