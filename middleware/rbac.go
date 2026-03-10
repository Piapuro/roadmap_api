package middleware

import (
	"github.com/labstack/echo/v4"
)

// RBAC checks whether the authenticated user has one of the required roles.
// TODO: implement after 3/7
func RBAC(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// NOTE: RBAC is not yet implemented. All requests pass through.
			// TODO: implement after JWT auth is ready (#007)
			_ = requiredRoles
			return next(c)
		}
	}
}
