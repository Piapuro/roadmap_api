package service

import "github.com/your-name/roadmap/api/adapter"

type RoadmapService struct {
	aiAdapter *adapter.AIAdapter
}

func NewRoadmapService(aiAdapter *adapter.AIAdapter) *RoadmapService {
	return &RoadmapService{aiAdapter: aiAdapter}
}
