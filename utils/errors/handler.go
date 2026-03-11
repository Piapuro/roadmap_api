package errors

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func NewGlobalErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var appErr *AppError
		if errors.As(err, &appErr) {
			_ = appErr.JSON(c)
			return
		}

		logger.Error("unexpected error", zap.Error(err))
		_ = c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    ErrInternal.Code,
			Message: ErrInternal.Message,
		})
	}
}
