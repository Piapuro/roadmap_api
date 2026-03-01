package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func BadRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, ErrorResponse{Code: http.StatusBadRequest, Message: msg})
}

func Unauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, ErrorResponse{Code: http.StatusUnauthorized, Message: "unauthorized"})
}

func NotFound(c echo.Context, msg string) error {
	return c.JSON(http.StatusNotFound, ErrorResponse{Code: http.StatusNotFound, Message: msg})
}

func InternalServerError(c echo.Context, msg string) error {
	return c.JSON(http.StatusInternalServerError, ErrorResponse{Code: http.StatusInternalServerError, Message: msg})
}
