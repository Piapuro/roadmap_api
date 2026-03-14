package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/Piapuro/roadmap_api/query"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TeamRole はチームスコープのロールレベルを表す。
// DB の team_roles.level に対応し、値が大きいほど上位権限。
type TeamRole int16

const (
	// TeamRoleMember はチームメンバー（team_role_id = 1）。
	TeamRoleMember TeamRole = 1
	// TeamRoleOwner はチームオーナー（team_role_id = 2）。
	TeamRoleOwner TeamRole = 2
)

const (
	// ContextKeyTeamRole は検証済みチームロールを格納するコンテキストキー。
	ContextKeyTeamRole = "team_role"
	// ContextKeyTeamID は検証済みチームIDを格納するコンテキストキー。
	ContextKeyTeamID = "team_id"
)

// TeamScopeAuth はチームスコープの権限チェックミドルウェア。
// JWT 検証（SupabaseAuth.Verify）の後に使用する。
type TeamScopeAuth struct {
	q *query.Queries
}

// NewTeamScopeAuth は TeamScopeAuth を生成する。
func NewTeamScopeAuth(q *query.Queries) *TeamScopeAuth {
	return &TeamScopeAuth{q: q}
}

// RequireMember はチームメンバー以上（MEMBER / OWNER）のみ通過させるミドルウェアを返す。
// パスパラメータ `:id` からチームIDを取得して権限を検証する。
func (m *TeamScopeAuth) RequireMember() echo.MiddlewareFunc {
	return m.requireRole(TeamRoleMember)
}

// RequireOwner はチームオーナーのみ通過させるミドルウェアを返す。
func (m *TeamScopeAuth) RequireOwner() echo.MiddlewareFunc {
	return m.requireRole(TeamRoleOwner)
}

// requireRole は minRole 以上のロールを持つユーザーのみ通過させるミドルウェアを返す。
func (m *TeamScopeAuth) requireRole(minRole TeamRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userIDStr, ok := c.Get(ContextKeyUserID).(string)
			if !ok || userIDStr == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
			}

			teamIDStr := c.Param("id")
			teamID, err := uuid.Parse(teamIDStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid team id"})
			}

			row, err := m.q.GetUserTeamRoleID(context.Background(), query.GetUserTeamRoleIDParams{
				UserID: userID,
				TeamID: teamID,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden: not a team member"})
				}
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}

			userRole := TeamRole(row.TeamRoleLevel)
			if userRole < minRole {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden: insufficient team role"})
			}

			// 検証済みロール情報をコンテキストにセット（後続ハンドラーで再クエリ不要）
			c.Set(ContextKeyTeamRole, string(row.TeamRoleName))
			c.Set(ContextKeyTeamID, teamID.String())

			return next(c)
		}
	}
}
