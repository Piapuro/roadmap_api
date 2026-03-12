package service_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Piapuro/roadmap_api/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClient は HTTPClient インターフェースのテスト用モック
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func newMockResponse(statusCode int, body interface{}) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(string(b))),
	}
}

func newAuthService(client *mockHTTPClient) *service.AuthService {
	return service.NewAuthService("https://example.supabase.co", "anon-key", client)
}

// --- Login テスト ---

func TestLogin_Success(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "/auth/v1/token", req.URL.Path)
			assert.Equal(t, "grant_type=password", req.URL.RawQuery)
			assert.Equal(t, "anon-key", req.Header.Get("apikey"))
			return newMockResponse(http.StatusOK, map[string]interface{}{
				"access_token":  "access-token-abc",
				"token_type":    "bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token-xyz",
				"user": map[string]interface{}{
					"id":            "user-uuid-123",
					"email":         "test@example.com",
					"created_at":    "2024-01-01T00:00:00Z",
					"user_metadata": map[string]interface{}{"full_name": "山田太郎"},
				},
			}), nil
		},
	}

	svc := newAuthService(mock)
	res, err := svc.Login(context.Background(), "test@example.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, "access-token-abc", res.AccessToken)
	assert.Equal(t, "bearer", res.TokenType)
	assert.Equal(t, 3600, res.ExpiresIn)
	assert.Equal(t, "refresh-token-xyz", res.RefreshToken)
	assert.Equal(t, "user-uuid-123", res.User.ID)
	assert.Equal(t, "test@example.com", res.User.Email)
	assert.Equal(t, "山田太郎", res.User.Name)
}

func TestLogin_UserMetadataNoName(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newMockResponse(http.StatusOK, map[string]interface{}{
				"access_token":  "token",
				"token_type":    "bearer",
				"expires_in":    3600,
				"refresh_token": "refresh",
				"user": map[string]interface{}{
					"id":            "uuid",
					"email":         "test@example.com",
					"created_at":    "2024-01-01T00:00:00Z",
					"user_metadata": map[string]interface{}{},
				},
			}), nil
		},
	}

	svc := newAuthService(mock)
	res, err := svc.Login(context.Background(), "test@example.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, "", res.User.Name) // full_name なし → 空文字
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newMockResponse(http.StatusBadRequest, map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Invalid login credentials",
			}), nil
		},
	}

	svc := newAuthService(mock)
	_, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")

	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestLogin_SupabaseServerError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newMockResponse(http.StatusInternalServerError, map[string]interface{}{
				"message": "internal server error",
			}), nil
		},
	}

	svc := newAuthService(mock)
	_, err := svc.Login(context.Background(), "test@example.com", "password123")

	require.Error(t, err)
	assert.NotErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestLogin_NetworkError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	svc := newAuthService(mock)
	_, err := svc.Login(context.Background(), "test@example.com", "password123")

	require.Error(t, err)
}

// --- Logout テスト ---

func TestLogout_Success(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "/auth/v1/logout", req.URL.Path)
			assert.Equal(t, "Bearer my-access-token", req.Header.Get("Authorization"))
			assert.Equal(t, "anon-key", req.Header.Get("apikey"))
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		},
	}

	svc := newAuthService(mock)
	err := svc.Logout(context.Background(), "my-access-token")

	require.NoError(t, err)
}

func TestLogout_Unauthorized(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newMockResponse(http.StatusUnauthorized, map[string]interface{}{
				"message": "Invalid token",
			}), nil
		},
	}

	svc := newAuthService(mock)
	err := svc.Logout(context.Background(), "expired-token")

	require.Error(t, err)
}

func TestLogout_NetworkError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	svc := newAuthService(mock)
	err := svc.Logout(context.Background(), "my-access-token")

	require.Error(t, err)
}
