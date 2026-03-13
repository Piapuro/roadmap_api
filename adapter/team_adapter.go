package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Piapuro/roadmap_api/query"
	"github.com/google/uuid"
)

type TeamAdapter struct {
	q *query.Queries
}

func NewTeamAdapter(q *query.Queries) *TeamAdapter {
	return &TeamAdapter{q: q}
}

func (a *TeamAdapter) IssueInviteToken(ctx context.Context, teamID uuid.UUID, token string, expiresAt time.Time) (query.Team, error) {
	team, err := a.q.IssueInviteToken(ctx, query.IssueInviteTokenParams{
		ID:                   teamID,
		InviteToken:          sql.NullString{String: token, Valid: true},
		InviteTokenExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return query.Team{}, fmt.Errorf("issue invite token: %w", err)
	}
	return team, nil
}

func (a *TeamAdapter) GetTeamByInviteToken(ctx context.Context, token string) (query.Team, error) {
	team, err := a.q.GetTeamByInviteToken(ctx, sql.NullString{String: token, Valid: true})
	if err != nil {
		return query.Team{}, fmt.Errorf("get team by invite token: %w", err)
	}
	return team, nil
}

func (a *TeamAdapter) IsTeamOwner(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) (bool, error) {
	isOwner, err := a.q.IsTeamOwner(ctx, query.IsTeamOwnerParams{UserID: userID, TeamID: teamID})
	if err != nil {
		return false, fmt.Errorf("check team owner: %w", err)
	}
	return isOwner, nil
}

func (a *TeamAdapter) IsTeamMember(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) (bool, error) {
	isMember, err := a.q.IsTeamMember(ctx, query.IsTeamMemberParams{UserID: userID, TeamID: teamID})
	if err != nil {
		return false, fmt.Errorf("check team member: %w", err)
	}
	return isMember, nil
}

func (a *TeamAdapter) JoinTeamAsMember(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) error {
	if err := a.q.JoinTeamAsMember(ctx, query.JoinTeamAsMemberParams{UserID: userID, TeamID: teamID}); err != nil {
		return fmt.Errorf("join team as member: %w", err)
	}
	return nil
}

func (a *TeamAdapter) ListTeamsByMember(ctx context.Context, userID uuid.UUID) ([]query.Team, error) {
	teams, err := a.q.ListTeamsByMember(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list teams by member: %w", err)
	}
	return teams, nil
}

func (a *TeamAdapter) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]query.ListTeamMembersRow, error) {
	members, err := a.q.ListTeamMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	return members, nil
}

func (a *TeamAdapter) ListUserSkills(ctx context.Context, userID uuid.UUID) ([]query.UserSkill, error) {
	skills, err := a.q.ListUserSkills(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user skills: %w", err)
	}
	return skills, nil
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
