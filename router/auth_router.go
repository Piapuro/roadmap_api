package router

import (
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/Piapuro/roadmap_api/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(e *echo.Echo, c *controller.AuthController, m *middleware.SupabaseAuth) {
	g := e.Group("/auth")
	g.POST("/signup", c.SignUp)
	g.POST("/login", c.Login)
	g.POST("/logout", c.Logout, m.Verify) // JWT必須
}
