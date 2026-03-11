package adapter

import (
	"context"
	"fmt"

	"github.com/Piapuro/roadmap_api/query"
)

type TeamAdapter struct {
	q *query.Queries
}

func NewTeamAdapter(q *query.Queries) *TeamAdapter {
	return &TeamAdapter{q: q}
}

func (a *TeamAdapter) CreateTeam(ctx context.Context, params query.CreateTeamParams) (query.Team, error) {
	team, err := a.q.CreateTeam(ctx, params)
	if err != nil {
		return query.Team{}, fmt.Errorf("create team: %w", err)
	}

	const teamOwnerRoleID = int16(2)
	if err := a.q.AssignTeamOwner(ctx, query.AssignTeamOwnerParams{
		UserID:     params.CreatedBy,
		TeamID:     team.ID,
		TeamRoleID: teamOwnerRoleID,
	}); err != nil {
		return query.Team{}, fmt.Errorf("assign team owner: %w", err)
	}

	return team, nil
}
