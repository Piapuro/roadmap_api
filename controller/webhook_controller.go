package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/labstack/echo/v4"
)

type WebhookController struct {
	webhookAdapter *adapter.WebhookAdapter
	secret         []byte
}

func NewWebhookController(webhookAdapter *adapter.WebhookAdapter, secret string) *WebhookController {
	return &WebhookController{webhookAdapter: webhookAdapter, secret: []byte(secret)}
}

func (wc *WebhookController) verifySignature(body []byte, sig string) bool {
	mac := hmac.New(sha256.New, wc.secret)
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
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
	var payload supabaseWebhookPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if payload.Type != "INSERT" || payload.Table != "users" {
		return c.JSON(http.StatusOK, map[string]string{"status": "ignored"})
	}

	if payload.Record.ID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing user id"})
	}

	if err := wc.webhookAdapter.AssignLoginUserRole(c.Request().Context(), payload.Record.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
