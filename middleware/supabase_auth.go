package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type SupabaseAuth struct {
	jwtSecret string
}

func NewSupabaseAuth(jwtSecret string) *SupabaseAuth {
	return &SupabaseAuth{jwtSecret: jwtSecret}
}

// Verify validates the Supabase JWT from the Authorization header.
func (m *SupabaseAuth) Verify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		_ = token
		// TODO: validate JWT with Supabase secret and set user context

		return next(c)
	}
}
