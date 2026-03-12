package router

import (
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/Piapuro/roadmap_api/middleware"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func RegisterAuthRoutes(e *echo.Echo, c *controller.AuthController, m *middleware.SupabaseAuth) {
	// ブルートフォース対策: 1分あたり20リクエストまで
	rateLimiter := echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(20))
	g := e.Group("/auth")
	g.POST("/signup", c.SignUp, rateLimiter)
	g.POST("/login", c.Login, rateLimiter)
	g.POST("/logout", c.Logout, m.Verify) // JWT必須
}
