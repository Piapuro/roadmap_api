package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/service"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// SignUp godoc
// @Summary      ユーザー登録
// @Description  メールアドレスとパスワードで新規ユーザーを登録します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.SignUpRequest  true  "登録情報"
// @Success      201   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string  "メールアドレスが既に登録済み"
// @Failure      500   {object}  map[string]string
// @Router       /auth/signup [post]
func (c *AuthController) SignUp(ctx echo.Context) error {
	var req requests.SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusCreated, nil)
}

// Login godoc
// @Summary      ログイン
// @Description  メールアドレスとパスワードで認証し、JWTトークンを取得します
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requests.LoginRequest  true  "ログイン情報"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx echo.Context) error {
	var req requests.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
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
	// TODO: implement
	return ctx.JSON(http.StatusOK, nil)
}
