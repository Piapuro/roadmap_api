package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

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

func (a *UserAdapter) GetMe(ctx context.Context, userID uuid.UUID) (query.User, error) {
	return a.q.GetUserByID(ctx, userID)
}

func (a *UserAdapter) UpdateMe(ctx context.Context, userID uuid.UUID, name string) (query.User, error) {
	return a.q.UpdateUserName(ctx, query.UpdateUserNameParams{
		ID:   userID,
		Name: name,
	})
}

func (a *UserAdapter) GetMySkills(ctx context.Context, userID uuid.UUID) (query.User, []query.UserSkill, error) {
	user, err := a.q.GetUserByID(ctx, userID)
	if err != nil {
		return query.User{}, nil, fmt.Errorf("get user: %w", err)
	}
	skills, err := a.q.ListUserSkills(ctx, userID)
	if err != nil {
		return query.User{}, nil, fmt.Errorf("list user skills: %w", err)
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
