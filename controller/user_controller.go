package controller

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/Piapuro/roadmap_api/query"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/response"
	"github.com/Piapuro/roadmap_api/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

// GetMe godoc
// @Summary      自分のプロフィール取得
// @Description  ログイン中のユーザー情報を返します
// @Tags         users
// @Produce      json
// @Success      200  {object}  response.UserResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /users/me [get]
func (c *UserController) GetMe(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

// UpdateMe godoc
// @Summary      自分のプロフィール更新
// @Description  ログイン中のユーザー情報を更新します
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      requests.UpdateUserRequest  true  "更新情報"
// @Success      200   {object}  response.UserResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /users/me [put]
func (c *UserController) UpdateMe(ctx echo.Context) error {
	var req requests.UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *UserController) GetMySkills(ctx echo.Context) error {
	userID, err := parseUserID(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	user, skills, err := c.userService.GetMySkills(ctx.Request().Context(), userID)
	if err != nil {
		// [C-1] ユーザーが存在しない場合は 404
		if errors.Is(err, sql.ErrNoRows) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, buildMySkillsResponse(user, skills))
}

func (c *UserController) UpsertMySkills(ctx echo.Context) error {
	userID, err := parseUserID(ctx)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	var req requests.UpsertSkillsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// [M-1] バリデーション実行
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// [M-2] upsert後の最新リソースを返す
	user, skills, err := c.userService.UpsertMySkills(ctx.Request().Context(), userID, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, buildMySkillsResponse(user, skills))
}

func parseUserID(ctx echo.Context) (uuid.UUID, error) {
	raw := ctx.Get("user_id")
	str, ok := raw.(string)
	if !ok {
		return uuid.UUID{}, echo.ErrUnauthorized
	}
	return uuid.Parse(str)
}

func buildMySkillsResponse(user query.User, skills []query.UserSkill) response.MySkillsResponse {
	skillResponses := make([]response.UserSkillResponse, len(skills))
	for i, s := range skills {
		var expYears *float64
		if s.ExperienceYears.Valid {
			f, err := strconv.ParseFloat(s.ExperienceYears.String, 64)
			if err == nil {
				expYears = &f
			}
		}
		skillResponses[i] = response.UserSkillResponse{
			ID:              s.ID.String(),
			SkillName:       s.SkillName,
			ExperienceYears: expYears,
			IsLearningGoal:  s.IsLearningGoal,
		}
	}

	bio := ""
	if user.Bio.Valid {
		bio = user.Bio.String
	}

	return response.MySkillsResponse{
		SkillLevel: user.SkillLevel,
		Bio:        bio,
		Skills:     skillResponses,
	}
}
