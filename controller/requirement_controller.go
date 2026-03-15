package controller

import (
	"errors"
	"net/http"

	"github.com/Piapuro/roadmap_api/middleware"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RequirementController struct {
	requirementService *service.RequirementService
}

func NewRequirementController(requirementService *service.RequirementService) *RequirementController {
	return &RequirementController{requirementService: requirementService}
}

// ListRequirements godoc
// @Summary      チームの要件定義一覧取得
// @Description  指定チームに属する要件定義の一覧を返します
// @Tags         requirements
// @Produce      json
// @Param        id   path      string  true  "チームID (UUID)"
// @Success      200  {array}   response.RequirementResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/{id}/requirements [get]
func (c *RequirementController) ListRequirements(ctx echo.Context) error {
	teamID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid team id"})
	}

	resp, err := c.requirementService.GetTeamRequirements(ctx.Request().Context(), teamID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return ctx.JSON(http.StatusOK, resp)
}

// CreateRequirement godoc
// @Summary      要件定義作成
// @Description  チームに新しい要件定義を作成します
// @Tags         requirements
// @Accept       json
// @Produce      json
// @Param        id    path      string                             true  "チームID (UUID)"
// @Param        body  body      requests.CreateRequirementRequest  true  "要件定義情報"
// @Success      201   {object}  response.RequirementResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /teams/{id}/requirements [post]
func (c *RequirementController) CreateRequirement(ctx echo.Context) error {
	teamID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid team id"})
	}

	var req requests.CreateRequirementRequest
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

	resp, err := c.requirementService.CreateRequirement(ctx.Request().Context(), userID, teamID, req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return ctx.JSON(http.StatusCreated, resp)
}

// GetRequirement godoc
// @Summary      要件定義取得
// @Description  指定IDの要件定義を返します
// @Tags         requirements
// @Produce      json
// @Param        id   path      string  true  "要件定義ID (UUID)"
// @Success      200  {object}  response.RequirementResponse
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /requirements/{id} [get]
func (c *RequirementController) GetRequirement(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid requirement id"})
	}

	resp, err := c.requirementService.GetRequirement(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "requirement not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return ctx.JSON(http.StatusOK, resp)
}

// UpdateRequirement godoc
// @Summary      要件定義更新
// @Description  指定IDの要件定義を更新します（draft 状態かつロードマップ未確定の場合のみ）
// @Tags         requirements
// @Accept       json
// @Produce      json
// @Param        id    path      string                             true  "要件定義ID (UUID)"
// @Param        body  body      requests.UpdateRequirementRequest  true  "更新情報"
// @Success      200   {object}  response.RequirementResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /requirements/{id} [put]
func (c *RequirementController) UpdateRequirement(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid requirement id"})
	}

	var req requests.UpdateRequirementRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	resp, err := c.requirementService.UpdateRequirement(ctx.Request().Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "requirement not found"})
		}
		if errors.Is(err, service.ErrRequirementLocked) {
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "requirement is locked and cannot be updated"})
		}
		if errors.Is(err, service.ErrRoadmapConfirmed) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "roadmap is confirmed, requirement cannot be edited"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return ctx.JSON(http.StatusOK, resp)
}

// SubmitRequirement godoc
// @Summary      要件定義を確定（ロック）
// @Description  要件定義のステータスを draft から locked へ遷移させます（ロードマップ未確定の場合のみ）
// @Tags         requirements
// @Produce      json
// @Param        id   path      string  true  "要件定義ID (UUID)"
// @Success      200  {object}  response.RequirementResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /requirements/{id}/submit [post]
func (c *RequirementController) SubmitRequirement(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid requirement id"})
	}

	resp, err := c.requirementService.LockRequirement(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "requirement not found"})
		}
		if errors.Is(err, service.ErrRequirementLocked) {
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "requirement is already locked"})
		}
		if errors.Is(err, service.ErrRoadmapConfirmed) {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "roadmap is confirmed, requirement cannot be edited"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return ctx.JSON(http.StatusOK, resp)
}
