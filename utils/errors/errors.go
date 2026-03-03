package errors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AppError struct {
	Status  int
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) JSON(c echo.Context) error {
	return c.JSON(e.Status, ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
	})
}

var (
	ErrUnauthorized = &AppError{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "認証が必要です"}
	ErrForbidden    = &AppError{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: "この操作の権限がありません"}
	ErrNotFound     = &AppError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "リソースが見つかりません"}
	ErrBadRequest   = &AppError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: "リクエストの形式が正しくありません"}
	ErrInternal     = &AppError{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "サーバーエラーが発生しました"}
)
