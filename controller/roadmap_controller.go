package controller

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
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
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
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
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
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
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
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
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
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
// @Param        id   path  string  true  "ロードマップID (UUID)"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /roadmaps/{id} [delete]
func (c *RoadmapController) DeleteRoadmap(ctx echo.Context) error {
	// TODO: implement
	return ctx.NoContent(http.StatusNoContent)
}

// SuggestMVP godoc
// @Summary      MVP範囲の自動提案
// @Description  要件定義データをもとにAIがMVP（最小実行可能製品）の機能リストを提案します
// @Tags         requirements
// @Produce      json
// @Param        id   path      string  true  "要件定義ID (UUID)"
// @Success      200  {object}  response.MVPSuggestionResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /requirements/{id}/suggest-mvp [post]
func (c *RoadmapController) SuggestMVP(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid requirement id"})
	}

	result, err := c.roadmapService.SuggestMVP(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrRequirementNotFound) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "requirement not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return ctx.JSON(http.StatusOK, result)
}
