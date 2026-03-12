package errors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	apperrors "github.com/Piapuro/roadmap_api/utils/errors"
	"go.uber.org/zap"
)

func newContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestErrUnauthorized(t *testing.T) {
	c, rec := newContext()
	_ = apperrors.ErrUnauthorized.JSON(c)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var res apperrors.ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Equal(t, "UNAUTHORIZED", res.Code)
}

func TestErrForbidden(t *testing.T) {
	c, rec := newContext()
	_ = apperrors.ErrForbidden.JSON(c)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	var res apperrors.ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Equal(t, "FORBIDDEN", res.Code)
}

func TestErrNotFound(t *testing.T) {
	c, rec := newContext()
	_ = apperrors.ErrNotFound.JSON(c)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	var res apperrors.ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Equal(t, "NOT_FOUND", res.Code)
}

func TestUnknownError_Returns500(t *testing.T) {
	e := echo.New()
	logger, _ := zap.NewDevelopment()
	handler := apperrors.NewGlobalErrorHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler(assert.AnError, c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var res apperrors.ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Equal(t, "INTERNAL_ERROR", res.Code)
}

func TestUnknownError_LogsError(t *testing.T) {
	e := echo.New()
	logger, _ := zap.NewDevelopment()
	handler := apperrors.NewGlobalErrorHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler(assert.AnError, c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
