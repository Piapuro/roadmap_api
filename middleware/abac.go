package middleware

import (
	"github.com/labstack/echo/v4"
)

// ABAC evaluates attribute-based access control rules after RBAC passes.
// TODO: implement after 3/7
func ABAC(action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: load ABAC rules from DB
			// TODO: evaluate condition_json against user/resource/env attributes
			// TODO: record result to abac_rule_logs
			_ = action
			return next(c)
		}
	}
}
