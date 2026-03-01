package service

import "github.com/your-name/roadmap/api/adapter"

type TeamService struct {
	teamAdapter *adapter.TeamAdapter
}

func NewTeamService(teamAdapter *adapter.TeamAdapter) *TeamService {
	return &TeamService{teamAdapter: teamAdapter}
}
