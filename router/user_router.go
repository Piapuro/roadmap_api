package router

import (
	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/controller"
	"github.com/your-name/roadmap/api/middleware"
)

func RegisterUserRoutes(e *echo.Echo, c *controller.UserController, m *middleware.SupabaseAuth) {
	g := e.Group("/users", m.Verify)
	g.GET("/me", c.GetMe)
	g.PUT("/me", c.UpdateMe)
}
