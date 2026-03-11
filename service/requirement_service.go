package service

import "github.com/Piapuro/roadmap_api/adapter"

type RequirementService struct {
	requirementAdapter *adapter.RequirementAdapter
}

func NewRequirementService(requirementAdapter *adapter.RequirementAdapter) *RequirementService {
	return &RequirementService{requirementAdapter: requirementAdapter}
}
