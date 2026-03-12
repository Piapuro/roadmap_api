package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const ContextKeyRoles = "roles"

// RBAC checks whether the authenticated user has one of the required roles.
// Role data must be set in the context under ContextKeyRoles ([]string) by a
// preceding middleware (e.g. after a database lookup). An empty requiredRoles
// list allows all authenticated requests through.
func RBAC(requiredRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if len(requiredRoles) == 0 {
				return next(c)
			}

			rawRoles := c.Get(ContextKeyRoles)
			if rawRoles == nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
			}

			userRoles, ok := rawRoles.([]string)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
			}

			roleSet := make(map[string]struct{}, len(userRoles))
			for _, r := range userRoles {
				roleSet[r] = struct{}{}
			}
			for _, required := range requiredRoles {
				if _, found := roleSet[required]; found {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
	}
}
