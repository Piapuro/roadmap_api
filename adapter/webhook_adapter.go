package adapter

import (
	"context"
	"database/sql"
	"fmt"
)

type WebhookAdapter struct {
	db *sql.DB
}

func NewWebhookAdapter(db *sql.DB) *WebhookAdapter {
	return &WebhookAdapter{db: db}
}

// mn-2: マジックナンバーを定数化
const globalRoleLoginUser = 2

// AssignLoginUserRole inserts a LOGIN_USER role for the given user.
func (a *WebhookAdapter) AssignLoginUserRole(ctx context.Context, userID string) error {
	const q = `
		INSERT INTO user_global_roles (user_id, global_role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	if _, err := a.db.ExecContext(ctx, q, userID, globalRoleLoginUser); err != nil {
		return fmt.Errorf("assign login user role: %w", err)
	}
	return nil
}
