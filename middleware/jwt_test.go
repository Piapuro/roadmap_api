package middleware_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/Piapuro/roadmap_api/middleware"
)

// makeNoneToken constructs a JWT with alg:none (no signature) for security testing.
func makeNoneToken(sub string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"` + sub + `","exp":9999999999}`))
	return header + "." + payload + "."
}

const testSecret = "test-secret-key"

func makeToken(t testing.TB, secret string, sub string, exp time.Time) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   sub,
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestVerify(t *testing.T) {
	auth := middleware.NewSupabaseAuth(testSecret, "")
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
			authHeader: "Bearer " + makeToken(t, testSecret, "user-uuid", time.Now().Add(time.Hour)),
			wantStatus: http.StatusOK,
		},
		{
			name:       "JWTなし",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "不正なJWT（署名が違う）",
			authHeader: "Bearer " + makeToken(t, "wrong-secret", "user-uuid", time.Now().Add(time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "期限切れJWT",
			authHeader: "Bearer " + makeToken(t, testSecret, "user-uuid", time.Now().Add(-time.Hour)),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "アルゴリズムnone攻撃",
			authHeader: "Bearer " + makeNoneToken("user-uuid"),
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
