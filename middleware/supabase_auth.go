package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

		claims := jwt.MapClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(m.jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		c.Set("user_id", claims["sub"])
		return next(c)
	}
}
