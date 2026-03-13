package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/Piapuro/roadmap_api/query"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/response"
	"github.com/google/uuid"
)

var (
	ErrNotTeamOwner       = errors.New("not team owner")
	ErrInviteTokenExpired = errors.New("invite token expired")
	ErrInviteTokenNotFound = errors.New("invite token not found")
	ErrAlreadyTeamMember  = errors.New("already team member")
)

type TeamService struct {
	teamAdapter *adapter.TeamAdapter
}

func NewTeamService(teamAdapter *adapter.TeamAdapter) *TeamService {
	return &TeamService{teamAdapter: teamAdapter}
}

func (s *TeamService) IssueInviteToken(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) (response.InviteTokenResponse, error) {
	isOwner, err := s.teamAdapter.IsTeamOwner(ctx, userID, teamID)
	if err != nil {
		return response.InviteTokenResponse{}, fmt.Errorf("check team owner: %w", err)
	}
	if !isOwner {
		return response.InviteTokenResponse{}, ErrNotTeamOwner
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return response.InviteTokenResponse{}, fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

	team, err := s.teamAdapter.IssueInviteToken(ctx, teamID, token, expiresAt)
	if err != nil {
		return response.InviteTokenResponse{}, err
	}

	return response.InviteTokenResponse{
		TeamID:    team.ID.String(),
		Token:     token,
		InviteURL: "/teams/join?token=" + token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *TeamService) JoinTeam(ctx context.Context, userID uuid.UUID, token string) (response.JoinTeamResponse, error) {
	team, err := s.teamAdapter.GetTeamByInviteToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.JoinTeamResponse{}, ErrInviteTokenNotFound
		}
		return response.JoinTeamResponse{}, fmt.Errorf("get team by invite token: %w", err)
	}

	if !team.InviteTokenExpiresAt.Valid || time.Now().UTC().After(team.InviteTokenExpiresAt.Time) {
		return response.JoinTeamResponse{}, ErrInviteTokenExpired
	}

	isMember, err := s.teamAdapter.IsTeamMember(ctx, userID, team.ID)
	if err != nil {
		return response.JoinTeamResponse{}, fmt.Errorf("check team member: %w", err)
	}
	if isMember {
		return response.JoinTeamResponse{}, ErrAlreadyTeamMember
	}

	if err := s.teamAdapter.JoinTeamAsMember(ctx, userID, team.ID); err != nil {
		return response.JoinTeamResponse{}, err
	}

	return response.JoinTeamResponse{
		TeamID:   team.ID.String(),
		UserID:   userID.String(),
		JoinedAt: time.Now().UTC(),
	}, nil
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
