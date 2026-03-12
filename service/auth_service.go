package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Piapuro/roadmap_api/response"
)

var (
	// ErrEmailAlreadyExists はメールアドレスが既に登録済みの場合に返されるエラー
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials はメールアドレスまたはパスワードが誤っている場合に返されるエラー
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// HTTPClient is an interface for making HTTP requests, enabling test injection.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AuthService struct {
	supabaseURL     string
	supabaseAnonKey string
	httpClient      HTTPClient
}

func NewAuthService(supabaseURL, supabaseAnonKey string, httpClient HTTPClient) *AuthService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &AuthService{
		supabaseURL:     supabaseURL,
		supabaseAnonKey: supabaseAnonKey,
		httpClient:      httpClient,
	}
}

// --- 内部リクエスト/レスポンス型 ---

type supabaseSignUpRequest struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type supabaseSignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// supabaseAuthResponse は SignUp / Login 共通のレスポンス型
type supabaseAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID           string                 `json:"id"`
		Email        string                 `json:"email"`
		CreatedAt    time.Time              `json:"created_at"`
		UserMetadata map[string]interface{} `json:"user_metadata"`
	} `json:"user"`
}

// --- SignUp ---

func (s *AuthService) SignUp(ctx context.Context, email, password, name string) (*response.SignUpResponse, error) {
	body, err := json.Marshal(supabaseSignUpRequest{
		Email:    email,
		Password: password,
		Data:     map[string]interface{}{"full_name": name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := strings.TrimSuffix(s.supabaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/v1/signup", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.supabaseAnonKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errBody); decodeErr != nil {
			log.Printf("[AuthService] Supabase signup error: status=%d (failed to decode body: %v)", resp.StatusCode, decodeErr)
			return nil, fmt.Errorf("supabase error: status %d", resp.StatusCode)
		}
		log.Printf("[AuthService] Supabase signup error: status=%d body=%v", resp.StatusCode, errBody)
		if code, ok := errBody["error_code"].(string); ok && code == "user_already_exists" {
			return nil, ErrEmailAlreadyExists
		}
		errMsg := errBody["msg"]
		if errMsg == nil {
			errMsg = errBody["message"]
		}
		return nil, fmt.Errorf("supabase error: %v", errMsg)
	}

	var sbResp supabaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&sbResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.SignUpResponse{
		AccessToken:  sbResp.AccessToken,
		TokenType:    sbResp.TokenType,
		ExpiresIn:    sbResp.ExpiresIn,
		RefreshToken: sbResp.RefreshToken,
		User: response.UserResponse{
			ID:        sbResp.User.ID,
			Email:     sbResp.User.Email,
			Name:      name,
			CreatedAt: sbResp.User.CreatedAt,
		},
	}, nil
}

// --- Login ---

func (s *AuthService) Login(ctx context.Context, email, password string) (*response.LoginResponse, error) {
	body, err := json.Marshal(supabaseSignInRequest{Email: email, Password: password})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := strings.TrimSuffix(s.supabaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/v1/token?grant_type=password", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.supabaseAnonKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errBody); decodeErr != nil {
			log.Printf("[AuthService] Supabase login error: status=%d (failed to decode body: %v)", resp.StatusCode, decodeErr)
			return nil, fmt.Errorf("supabase error: status %d", resp.StatusCode)
		}
		log.Printf("[AuthService] Supabase login error: status=%d body=%v", resp.StatusCode, errBody)
		if errCode, ok := errBody["error"].(string); ok && errCode == "invalid_grant" {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("supabase error: status %d", resp.StatusCode)
	}

	var sbResp supabaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&sbResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	name, _ := sbResp.User.UserMetadata["full_name"].(string)

	return &response.LoginResponse{
		AccessToken:  sbResp.AccessToken,
		TokenType:    sbResp.TokenType,
		ExpiresIn:    sbResp.ExpiresIn,
		RefreshToken: sbResp.RefreshToken,
		User: response.UserResponse{
			ID:        sbResp.User.ID,
			Email:     sbResp.User.Email,
			Name:      name,
			CreatedAt: sbResp.User.CreatedAt,
		},
	}, nil
}

// --- Logout ---

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	baseURL := strings.TrimSuffix(s.supabaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/v1/logout", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("apikey", s.supabaseAnonKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase logout error: status %d", resp.StatusCode)
	}
	return nil
}
