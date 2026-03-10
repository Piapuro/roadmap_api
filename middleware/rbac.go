package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RBAC checks whether the authenticated user has one of the required roles.
// TODO: implement - role extraction from context and validation against requiredRoles
func RBAC(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Fail-closed until implementation is complete.
			_ = requiredRoles
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
	}
}
