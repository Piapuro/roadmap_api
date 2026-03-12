package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/Piapuro/roadmap_api/query"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/response"
	"github.com/google/uuid"
)

type TeamService struct {
	teamAdapter *adapter.TeamAdapter
}

func NewTeamService(teamAdapter *adapter.TeamAdapter) *TeamService {
	return &TeamService{teamAdapter: teamAdapter}
}

func (s *TeamService) CreateTeam(ctx context.Context, userID uuid.UUID, req requests.CreateTeamRequest) (response.TeamResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return response.TeamResponse{}, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return response.TeamResponse{}, fmt.Errorf("invalid end_date: %w", err)
	}

	team, err := s.teamAdapter.CreateTeam(ctx, query.CreateTeamParams{
		Name:      req.Name,
		Goal:      req.Goal,
		Level:     req.Level,
		StartDate: startDate,
		EndDate:   endDate,
		CreatedBy: userID,
	})
	if err != nil {
		return response.TeamResponse{}, err
	}

	return response.TeamResponse{
		ID:        team.ID.String(),
		Name:      team.Name,
		Goal:      team.Goal,
		Level:     team.Level,
		StartDate: team.StartDate,
		EndDate:   team.EndDate,
		CreatedBy: team.CreatedBy.String(),
		CreatedAt: team.CreatedAt,
	}, nil
}
