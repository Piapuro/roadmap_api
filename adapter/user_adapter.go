package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/Piapuro/roadmap_api/query"
	"github.com/google/uuid"
)

type UserAdapter struct {
	q  *query.Queries
	db *sql.DB
}

func NewUserAdapter(q *query.Queries, db *sql.DB) *UserAdapter {
	return &UserAdapter{q: q, db: db}
}

// EnsureUserExists は Supabase Auth で認証したユーザーを user_profiles に同期する。
// ローカル DB 使用時、auth.users のトリガーが動かないため SignUp/Login 後に呼ぶ。
func (a *UserAdapter) EnsureUserExists(ctx context.Context, userID uuid.UUID, name string) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qtx := a.q.WithTx(tx)
	if err := qtx.EnsureUser(ctx, query.EnsureUserParams{ID: userID, Name: truncateName(name)}); err != nil {
		return fmt.Errorf("ensure user: %w", err)
	}
	if err := qtx.AssignGlobalRole(ctx, query.AssignGlobalRoleParams{UserID: userID, GlobalRoleID: int16(globalRoleLoginUser)}); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return tx.Commit()
}

func truncateName(name string) string {
	const maxLen = 20
	if utf8.RuneCountInString(name) <= maxLen {
		return name
	}
	return string([]rune(name)[:maxLen])
}

func (a *UserAdapter) GetMe(ctx context.Context, userID uuid.UUID) (query.UserProfile, error) {
	return a.q.GetUserByID(ctx, userID)
}

func (a *UserAdapter) UpdateMe(ctx context.Context, userID uuid.UUID, name string) (query.UserProfile, error) {
	return a.q.UpdateUserName(ctx, query.UpdateUserNameParams{
		ID:   userID,
		Name: name,
	})
}

func (a *UserAdapter) GetMySkills(ctx context.Context, userID uuid.UUID) (query.UserProfile, []query.UserSkill, error) {
	user, err := a.q.GetUserByID(ctx, userID)
	if err != nil {
		return query.UserProfile{}, nil, fmt.Errorf("get user: %w", err)
	}
	skills, err := a.q.ListUserSkills(ctx, userID)
	if err != nil {
		return query.UserProfile{}, nil, fmt.Errorf("list user skills: %w", err)
	}
	return user, skills, nil
}

type SkillInput struct {
	SkillName       string
	ExperienceYears *float64
	IsLearningGoal  bool
}

func (a *UserAdapter) UpsertSkills(ctx context.Context, userID uuid.UUID, skillLevel, bio string, skills []SkillInput) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qtx := a.q.WithTx(tx)

	if err := qtx.DeleteUserSkills(ctx, userID); err != nil {
		return fmt.Errorf("delete user skills: %w", err)
	}

	for _, s := range skills {
		var expYears sql.NullString
		if s.ExperienceYears != nil {
			expYears = sql.NullString{
				String: strconv.FormatFloat(*s.ExperienceYears, 'f', 1, 64),
				Valid:  true,
			}
		}
		if _, err := qtx.CreateUserSkill(ctx, query.CreateUserSkillParams{
			UserID:          userID,
			SkillName:       s.SkillName,
			ExperienceYears: expYears,
			IsLearningGoal:  s.IsLearningGoal,
		}); err != nil {
			return fmt.Errorf("create user skill: %w", err)
		}
	}

	if _, err := qtx.UpdateUserProfile(ctx, query.UpdateUserProfileParams{
		ID:         userID,
		SkillLevel: skillLevel,
		Bio:        sql.NullString{String: bio, Valid: bio != ""},
	}); err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	return tx.Commit()
}
