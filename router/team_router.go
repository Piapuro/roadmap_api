package router

import (
	"github.com/Piapuro/roadmap_api/controller"
	"github.com/Piapuro/roadmap_api/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterTeamRoutes(e *echo.Echo, c *controller.TeamController, rc *controller.RequirementController, m *middleware.SupabaseAuth, ts *middleware.TeamScopeAuth) {
	g := e.Group("/teams", m.Verify)

	// 認証済みユーザーであれば誰でもアクセス可能
	g.POST("", c.CreateTeam)
	g.GET("", c.GetTeams)
	g.POST("/join", c.JoinTeam)

	// チームメンバー以上が必要
	g.GET("/:id", c.GetTeam, ts.RequireMember())
	g.GET("/:id/members", c.GetTeamMembers, ts.RequireMember())
	g.GET("/:id/requirements", rc.ListRequirements, ts.RequireMember())
	g.POST("/:id/requirements", rc.CreateRequirement, ts.RequireMember())

	// チームオーナーのみ
	g.PUT("/:id", c.UpdateTeam, ts.RequireOwner())
	g.DELETE("/:id", c.DeleteTeam, ts.RequireOwner())
	g.POST("/:id/invite", c.IssueInviteToken, ts.RequireOwner())
}
