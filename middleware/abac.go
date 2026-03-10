package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ABAC evaluates attribute-based access control rules after RBAC passes.
// TODO: implement - load rules from DB, evaluate condition_json, record to abac_rule_logs
func ABAC(action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Fail-closed until implementation is complete.
			_ = action
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
	}
}
