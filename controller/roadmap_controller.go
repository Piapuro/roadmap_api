package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/requests"
	"github.com/your-name/roadmap/api/service"
)

type RoadmapController struct {
	roadmapService *service.RoadmapService
}

func NewRoadmapController(roadmapService *service.RoadmapService) *RoadmapController {
	return &RoadmapController{roadmapService: roadmapService}
}

// CreateRoadmap godoc
// @Summary      ロードマップ作成
// @Description  新しいロードマップを作成します
// @Tags         roadmaps
// @Accept       json
// @Produce      json
// @Param        body  body      requests.CreateRoadmapRequest  true  "ロードマップ情報"
// @Success      201   {object}  response.RoadmapResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Security     BearerAuth
// @Router       /roadmaps [post]
func (c *RoadmapController) CreateRoadmap(ctx echo.Context) error {
	var req requests.CreateRoadmapRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusCreated, nil)
}

// GetRoadmaps godoc
// @Summary      ロードマップ一覧取得
// @Description  ログインユーザーが閲覧可能なロードマップ一覧を返します
// @Tags         roadmaps
// @Produce      json
// @Success      200  {array}   response.RoadmapResponse
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /roadmaps [get]
func (c *RoadmapController) GetRoadmaps(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// GetRoadmap godoc
// @Summary      ロードマップ取得
// @Description  指定IDのロードマップを返します
// @Tags         roadmaps
// @Produce      json
// @Param        id   path      string  true  "ロードマップID (UUID)"
// @Success      200  {object}  response.RoadmapResponse
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /roadmaps/{id} [get]
func (c *RoadmapController) GetRoadmap(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// UpdateRoadmap godoc
// @Summary      ロードマップ更新
// @Description  指定IDのロードマップを更新します
// @Tags         roadmaps
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "ロードマップID (UUID)"
// @Param        body  body      requests.UpdateRoadmapRequest  true  "更新情報"
// @Success      200   {object}  response.RoadmapResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Security     BearerAuth
// @Router       /roadmaps/{id} [put]
func (c *RoadmapController) UpdateRoadmap(ctx echo.Context) error {
	var req requests.UpdateRoadmapRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// DeleteRoadmap godoc
// @Summary      ロードマップ削除
// @Description  指定IDのロードマップを削除します
// @Tags         roadmaps
// @Produce      json
// @Param        id   path  string  true  "ロードマップID (UUID)"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /roadmaps/{id} [delete]
func (c *RoadmapController) DeleteRoadmap(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusNoContent, nil)
}
