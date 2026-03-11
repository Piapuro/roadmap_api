package controller

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/labstack/echo/v4"
)

type WebhookController struct {
	webhookAdapter *adapter.WebhookAdapter
	secret         string
}

// [C-1] 空シークレットはサーバー起動時に弾く
func NewWebhookController(webhookAdapter *adapter.WebhookAdapter, secret string) (*WebhookController, error) {
	if secret == "" {
		return nil, fmt.Errorf("WEBHOOK_SECRET must not be empty")
	}
	return &WebhookController{webhookAdapter: webhookAdapter, secret: secret}, nil
}

func (wc *WebhookController) verifySecret(incoming string) bool {
	return subtle.ConstantTimeCompare([]byte(incoming), []byte(wc.secret)) == 1
}

type supabaseWebhookPayload struct {
	Type   string `json:"type"`
	Schema string `json:"schema"`
	Table  string `json:"table"`
	Record struct {
		ID string `json:"id"`
	} `json:"record"`
}

// OnUserCreated receives a Supabase database webhook on auth.users INSERT
// and assigns the LOGIN_USER global role to the new user.
func (wc *WebhookController) OnUserCreated(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
	}

	sig := c.Request().Header.Get("x-webhook-secret")
	if !wc.verifySecret(sig) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
	}

	var payload supabaseWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	// [C-2] schema も含めて検証
	if payload.Type != "INSERT" || payload.Schema != "auth" || payload.Table != "users" {
		return c.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}

	if payload.Record.ID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing user id"})
	}

	if err := wc.webhookAdapter.AssignLoginUserRole(c.Request().Context(), payload.Record.ID); err != nil {
		// [C-3] 内部エラー詳細を外部に漏らさない
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
