package router

import (
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/controller"
)

func RegisterWebhookRoutes(e *echo.Echo, c *controller.WebhookController) {
	g := e.Group("/webhooks")
	g.POST("/supabase/user-created", c.OnUserCreated)
}
