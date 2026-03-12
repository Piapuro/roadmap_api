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

// mn-1: ヘッダーキーを定数化
const webhookSecretHeader = "x-webhook-secret"

// mn-4: Record を名前付き型に分離
type webhookRecord struct {
	ID string `json:"id"`
}

type supabaseWebhookPayload struct {
	Type   string        `json:"type"`
	Schema string        `json:"schema"`
	Table  string        `json:"table"`
	Record webhookRecord `json:"record"`
}

type WebhookController struct {
	webhookAdapter *adapter.WebhookAdapter
	secret         string
}

func NewWebhookController(webhookAdapter *adapter.WebhookAdapter, secret string) (*WebhookController, error) {
	if secret == "" {
		return nil, fmt.Errorf("WEBHOOK_SECRET must not be empty")
	}
	return &WebhookController{webhookAdapter: webhookAdapter, secret: secret}, nil
}

func (wc *WebhookController) verifySecret(incoming string) bool {
	return subtle.ConstantTimeCompare([]byte(incoming), []byte(wc.secret)) == 1
}

// OnUserCreated receives a Supabase database webhook on auth.users INSERT
// and assigns the LOGIN_USER global role to the new user.
//
// @Summary      ユーザー作成Webhook
// @Description  Supabase auth.users INSERT 時に LOGIN_USER グローバルロールを付与する
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        x-webhook-secret  header    string  true  "Webhook署名シークレット"
// @Success      200               {object}  map[string]string
// @Failure      400               {object}  map[string]string
// @Failure      401               {object}  map[string]string
// @Failure      500               {object}  map[string]string
// @Router       /webhooks/supabase/user-created [post]
func (wc *WebhookController) OnUserCreated(c echo.Context) error {
	// M-2: Fail Fast — ボディ読み込み前にシークレット検証
	sig := c.Request().Header.Get(webhookSecretHeader)
	if !wc.verifySecret(sig) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
	}

	// M-1: ボディサイズ上限 1MB
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, 1<<20))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
	}

	var payload supabaseWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	if payload.Type != "INSERT" || payload.Schema != "auth" || payload.Table != "users" {
		return c.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}

	if payload.Record.ID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing user id"})
	}

	if err := wc.webhookAdapter.AssignLoginUserRole(c.Request().Context(), payload.Record.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
