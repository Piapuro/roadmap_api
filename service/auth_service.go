package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Piapuro/roadmap_api/response"
)

// ErrEmailAlreadyExists はメールアドレスが既に登録済みの場合に返されるエラー
var ErrEmailAlreadyExists = errors.New("email already exists")

type AuthService struct {
	supabaseURL     string
	supabaseAnonKey string
}

func NewAuthService(supabaseURL, supabaseAnonKey string) *AuthService {
	return &AuthService{
		supabaseURL:     supabaseURL,
		supabaseAnonKey: supabaseAnonKey,
	}
}

type supabaseSignUpRequest struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type supabaseSignUpResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"user"`
}

func (s *AuthService) SignUp(email, password, name string) (*response.SignUpResponse, error) {
	body, err := json.Marshal(supabaseSignUpRequest{
		Email:    email,
		Password: password,
		Data:     map[string]interface{}{"full_name": name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.supabaseURL+"/auth/v1/signup", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.supabaseAnonKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		if code, ok := errBody["error_code"].(string); ok && code == "user_already_exists" {
			return nil, ErrEmailAlreadyExists
		}
		errMsg := errBody["msg"]
		if errMsg == nil {
			errMsg = errBody["message"]
		}
		return nil, fmt.Errorf("supabase error: %v", errMsg)
	}

	var sbResp supabaseSignUpResponse
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
