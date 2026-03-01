package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RBAC checks whether the authenticated user has one of the required roles.
// TODO: implement after 3/7
func RBAC(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: extract role from context set by SupabaseAuth middleware
			// TODO: check against requiredRoles
			_ = requiredRoles
			if false {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
			}
			return next(c)
		}
	}
}
