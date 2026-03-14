package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthController(authService *service.AuthService, userService *service.UserService) *AuthController {
	return &AuthController{authService: authService, userService: userService}
}

// SignUp godoc
// @Summary      ユーザー登録
// @Description  メールアドレスとパスワードで新規ユーザーを登録します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.SignUpRequest  true  "登録情報"
// @Success      201   {object}  response.SignUpResponse
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string  "メールアドレスが既に登録済み"
// @Failure      500   {object}  map[string]string
// @Router       /auth/signup [post]
func (c *AuthController) SignUp(ctx echo.Context) error {
	var req requests.SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	res, err := c.authService.SignUp(ctx.Request().Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			return ctx.JSON(http.StatusConflict, map[string]string{"error": "email already registered"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	userID, err := uuid.Parse(res.User.ID)
	if err != nil {
		ctx.Logger().Errorf("SignUp: invalid user ID %q: %v", res.User.ID, err)
	} else if err := c.userService.EnsureUserExists(ctx.Request().Context(), userID, req.Name, req.Email); err != nil {
		ctx.Logger().Errorf("SignUp: EnsureUserExists failed for user %s: %v", userID, err)
	}
	return ctx.JSON(http.StatusCreated, res)
}

// Login godoc
// @Summary      ログイン
// @Description  メールアドレスとパスワードで認証し、JWTトークンを取得します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.LoginRequest  true  "ログイン情報"
// @Success      200   {object}  response.LoginResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx echo.Context) error {
	var req requests.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	res, err := c.authService.Login(ctx.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	userID, err := uuid.Parse(res.User.ID)
	if err != nil {
		ctx.Logger().Errorf("Login: invalid user ID %q: %v", res.User.ID, err)
	} else if err := c.userService.EnsureUserExists(ctx.Request().Context(), userID, res.User.Name, req.Email); err != nil {
		ctx.Logger().Errorf("Login: EnsureUserExists failed for user %s: %v", userID, err)
	}
	return ctx.JSON(http.StatusOK, res)
}

// Logout godoc
// @Summary      ログアウト
// @Description  現在のセッションを無効化します
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (c *AuthController) Logout(ctx echo.Context) error {
	authHeader := ctx.Request().Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if err := c.authService.Logout(ctx.Request().Context(), token); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}
