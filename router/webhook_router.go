package router

import (
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func RegisterWebhookRoutes(e *echo.Echo, c *controller.WebhookController) {
	// M-4: 認証なし公開エンドポイントのため レート制限を適用（10 req/s）
	g := e.Group("/webhooks", echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(10)))
	g.POST("/supabase/user-created", c.OnUserCreated)
}
