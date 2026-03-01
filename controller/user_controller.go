package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/requests"
	"github.com/your-name/roadmap/api/service"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) GetMe(ctx echo.Context) error {
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}

func (c *UserController) UpdateMe(ctx echo.Context) error {
	var req requests.UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}
