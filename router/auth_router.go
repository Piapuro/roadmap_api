package router

import (
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/controller"
)

func RegisterAuthRoutes(e *echo.Echo, c *controller.AuthController) {
	g := e.Group("/auth")
	g.POST("/signup", c.SignUp)
	g.POST("/login", c.Login)
	g.POST("/logout", c.Logout)
}
