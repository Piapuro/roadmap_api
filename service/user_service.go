package service

import (
	"context"
	"strings"

	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/Piapuro/roadmap_api/query"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/google/uuid"
)

type UserService struct {
	userAdapter *adapter.UserAdapter
}

func NewUserService(userAdapter *adapter.UserAdapter) *UserService {
	return &UserService{userAdapter: userAdapter}
}

func (s *UserService) EnsureUserExists(ctx context.Context, userID uuid.UUID, name, email string) error {
	resolved := name
	if resolved == "" {
		resolved = strings.SplitN(email, "@", 2)[0]
	}
	return s.userAdapter.EnsureUserExists(ctx, userID, resolved)
}

func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (query.UserProfile, error) {
	return s.userAdapter.GetMe(ctx, userID)
}

func (s *UserService) UpdateMe(ctx context.Context, userID uuid.UUID, name string) (query.UserProfile, error) {
	return s.userAdapter.UpdateMe(ctx, userID, name)
}

func (s *UserService) GetMySkills(ctx context.Context, userID uuid.UUID) (query.UserProfile, []query.UserSkill, error) {
	return s.userAdapter.GetMySkills(ctx, userID)
}

func (s *UserService) UpsertMySkills(ctx context.Context, userID uuid.UUID, req requests.UpsertSkillsRequest) (query.UserProfile, []query.UserSkill, error) {
	skills := make([]adapter.SkillInput, len(req.Skills))
	for i, s := range req.Skills {
		skills[i] = adapter.SkillInput{
			SkillName:       s.SkillName,
			ExperienceYears: s.ExperienceYears,
			IsLearningGoal:  s.IsLearningGoal,
		}
	}
	if err := s.userAdapter.UpsertSkills(ctx, userID, req.SkillLevel, req.Bio, skills); err != nil {
		return query.UserProfile{}, nil, err
	}
	return s.userAdapter.GetMySkills(ctx, userID)
}
