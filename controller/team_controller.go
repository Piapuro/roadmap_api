package controller

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/middleware"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
)

type TeamController struct {
	teamService *service.TeamService
}

func NewTeamController(teamService *service.TeamService) *TeamController {
	return &TeamController{teamService: teamService}
}

// CreateTeam godoc
// @Summary      チーム作成
// @Description  新しいチームを作成します
// @Tags         teams
// @Accept       json
// @Produce      json
// @Param        body  body      requests.CreateTeamRequest  true  "チーム情報"
// @Success      201   {object}  response.TeamResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams [post]
func (c *TeamController) CreateTeam(ctx echo.Context) error {
	var req requests.CreateTeamRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	userIDStr, ok := ctx.Get(middleware.ContextKeyUserID).(string)
	if !ok || userIDStr == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	resp, err := c.teamService.CreateTeam(ctx.Request().Context(), userID, req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return ctx.JSON(http.StatusCreated, resp)
}

// GetTeams godoc
// @Summary      チーム一覧取得
// @Description  ログインユーザーが所属するチームの一覧を返します
// @Tags         teams
// @Produce      json
// @Success      200  {array}   response.TeamResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams [get]
func (c *TeamController) GetTeams(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// GetTeam godoc
// @Summary      チーム取得
// @Description  指定IDのチーム情報を返します
// @Tags         teams
// @Produce      json
// @Param        id   path      string  true  "チームID (UUID)"
// @Success      200  {object}  response.TeamResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/{id} [get]
func (c *TeamController) GetTeam(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// UpdateTeam godoc
// @Summary      チーム更新
// @Description  指定IDのチーム情報を更新します
// @Tags         teams
// @Accept       json
// @Produce      json
// @Param        id    path      string                      true  "チームID (UUID)"
// @Param        body  body      requests.UpdateTeamRequest  true  "更新情報"
// @Success      200   {object}  response.TeamResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/{id} [put]
func (c *TeamController) UpdateTeam(ctx echo.Context) error {
	var req requests.UpdateTeamRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// GetTeamMembers godoc
// @Summary      チームメンバー一覧取得
// @Description  チームに所属するメンバーのロール・スキル情報を返します（チームメンバーのみアクセス可能）
// @Tags         teams
// @Produce      json
// @Param        id   path      string  true  "チームID (UUID)"
// @Success      200  {array}   response.TeamMemberResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/{id}/members [get]
func (c *TeamController) GetTeamMembers(ctx echo.Context) error {
	userIDStr, ok := ctx.Get(middleware.ContextKeyUserID).(string)
	if !ok || userIDStr == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	teamID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid team id"})
	}

	members, err := c.teamService.GetTeamMembers(ctx.Request().Context(), userID, teamID)
	if err != nil {
		if errors.Is(err, service.ErrNotTeamMember) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "access denied: not a team member"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return ctx.JSON(http.StatusOK, members)
}

// IssueInviteToken godoc
// @Summary      招待トークン発行
// @Description  チームへの招待トークンを発行します（チームオーナーのみ）
// @Tags         teams
// @Produce      json
// @Param        id   path      string  true  "チームID (UUID)"
// @Success      200  {object}  response.InviteTokenResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/{id}/invite [post]
func (c *TeamController) IssueInviteToken(ctx echo.Context) error {
	userIDStr, ok := ctx.Get(middleware.ContextKeyUserID).(string)
	if !ok || userIDStr == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	teamID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid team id"})
	}

	resp, err := c.teamService.IssueInviteToken(ctx.Request().Context(), userID, teamID)
	if err != nil {
		if errors.Is(err, service.ErrNotTeamOwner) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "only team owner can issue invite tokens"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return ctx.JSON(http.StatusOK, resp)
}

// JoinTeam godoc
// @Summary      招待リンクからチーム参加
// @Description  招待トークンを使用してチームに参加します
// @Tags         teams
// @Accept       json
// @Produce      json
// @Param        body  body      requests.JoinTeamRequest  true  "招待トークン"
// @Success      200   {object}  response.JoinTeamResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/join [post]
func (c *TeamController) JoinTeam(ctx echo.Context) error {
	var req requests.JoinTeamRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	userIDStr, ok := ctx.Get(middleware.ContextKeyUserID).(string)
	if !ok || userIDStr == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	resp, err := c.teamService.JoinTeam(ctx.Request().Context(), userID, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInviteTokenNotFound):
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid invite token"})
		case errors.Is(err, service.ErrInviteTokenExpired):
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invite token has expired"})
		case errors.Is(err, service.ErrAlreadyTeamMember):
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "already a team member"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return ctx.JSON(http.StatusOK, resp)
}

// DeleteTeam godoc
// @Summary      チーム削除
// @Description  指定IDのチームを削除します
// @Tags         teams
// @Param        id   path  string  true  "チームID (UUID)"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/{id} [delete]
func (c *TeamController) DeleteTeam(ctx echo.Context) error {
	// TODO: implement
	return ctx.NoContent(http.StatusNoContent)
}
