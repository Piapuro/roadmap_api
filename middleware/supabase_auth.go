package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const ContextKeyUserID = "user_id"

type SupabaseAuth struct {
	jwtSecret []byte
}

func NewSupabaseAuth(jwtSecret string) *SupabaseAuth {
	return &SupabaseAuth{jwtSecret: []byte(jwtSecret)}
}

type supabaseClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Verify validates the Supabase JWT and sets claims["sub"] as user_id in context.
func (m *SupabaseAuth) Verify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &supabaseClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.ErrUnauthorized
			}
			return m.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		sub, err := claims.GetSubject()
		if err != nil || sub == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		c.Set(ContextKeyUserID, sub)
		return next(c)
	}
}
