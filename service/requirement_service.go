package service

import "github.com/your-name/roadmap/api/adapter"

type RequirementService struct {
	requirementAdapter *adapter.RequirementAdapter
}

func NewRequirementService(requirementAdapter *adapter.RequirementAdapter) *RequirementService {
	return &RequirementService{requirementAdapter: requirementAdapter}
}
