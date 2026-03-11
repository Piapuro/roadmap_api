package service

import (
	"context"

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

func (s *UserService) GetMySkills(ctx context.Context, userID uuid.UUID) (query.User, []query.UserSkill, error) {
	return s.userAdapter.GetMySkills(ctx, userID)
}

func (s *UserService) UpsertMySkills(ctx context.Context, userID uuid.UUID, req requests.UpsertSkillsRequest) (query.User, []query.UserSkill, error) {
	skills := make([]adapter.SkillInput, len(req.Skills))
	for i, s := range req.Skills {
		skills[i] = adapter.SkillInput{
			SkillName:       s.SkillName,
			ExperienceYears: s.ExperienceYears,
			IsLearningGoal:  s.IsLearningGoal,
		}
	}
	if err := s.userAdapter.UpsertSkills(ctx, userID, req.SkillLevel, req.Bio, skills); err != nil {
		return query.User{}, nil, err
	}
	return s.userAdapter.GetMySkills(ctx, userID)
}
