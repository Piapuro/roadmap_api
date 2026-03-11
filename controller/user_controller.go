package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
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
