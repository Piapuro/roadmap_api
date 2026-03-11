package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
)

type RequirementController struct {
	requirementService *service.RequirementService
}

func NewRequirementController(requirementService *service.RequirementService) *RequirementController {
	return &RequirementController{requirementService: requirementService}
}

// CreateRequirement godoc
// @Summary      要件定義作成
// @Description  新しい要件定義を作成します
// @Tags         requirements
// @Accept       json
// @Produce      json
// @Param        body  body      requests.CreateRequirementRequest  true  "要件定義情報"
// @Success      201   {object}  response.RequirementResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /requirements [post]
func (c *RequirementController) CreateRequirement(ctx echo.Context) error {
	var req requests.CreateRequirementRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusCreated, nil)
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
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// UpdateRequirement godoc
// @Summary      要件定義更新
// @Description  指定IDの要件定義を更新します
// @Tags         requirements
// @Accept       json
// @Produce      json
// @Param        id    path      string                             true  "要件定義ID (UUID)"
// @Param        body  body      requests.UpdateRequirementRequest  true  "更新情報"
// @Success      200   {object}  response.RequirementResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /requirements/{id} [put]
func (c *RequirementController) UpdateRequirement(ctx echo.Context) error {
	var req requests.UpdateRequirementRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// SubmitRequirement godoc
// @Summary      要件定義を提出
// @Description  要件定義のステータスを draft から submitted へ遷移させます
// @Tags         requirements
// @Produce      json
// @Param        id   path      string  true  "要件定義ID (UUID)"
// @Success      200  {object}  response.RequirementResponse
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string  "すでに提出済みの場合"
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /requirements/{id}/submit [post]
func (c *RequirementController) SubmitRequirement(ctx echo.Context) error {
	// TODO: implement（draft → submitted へのステータス遷移）
	return ctx.JSON(http.StatusOK, nil)
}
