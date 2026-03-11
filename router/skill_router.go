package router

import (
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/labstack/echo/v4"
)

func RegisterSkillRoutes(e *echo.Echo, c *controller.SkillController) {
	e.GET("/skills", c.ListSkillTags)
}
