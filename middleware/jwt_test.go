package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/your-name/roadmap/api/middleware"
)

const testSecret = "test-secret-key"

func makeToken(secret string, sub string, exp time.Time) string {
	claims := jwt.RegisteredClaims{
		Subject:   sub,
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestVerify(t *testing.T) {
	auth := middleware.NewSupabaseAuth(testSecret)
	handler := auth.Verify(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "正常なJWT",
			authHeader: "Bearer " + makeToken(testSecret, "user-uuid", time.Now().Add(time.Hour)),
			wantStatus: http.StatusOK,
		},
		{
			name:       "JWTなし",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "不正なJWT（署名が違う）",
			authHeader: "Bearer " + makeToken("wrong-secret", "user-uuid", time.Now().Add(time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "期限切れJWT",
			authHeader: "Bearer " + makeToken(testSecret, "user-uuid", time.Now().Add(-time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = handler(c)

			if rec.Code != tt.wantStatus {
				t.Errorf("got %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
