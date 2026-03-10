package service

import "github.com/Piapuro/roadmap_api/adapter"

type RoadmapService struct {
	aiAdapter *adapter.AIAdapter
}

func NewRoadmapService(aiAdapter *adapter.AIAdapter) *RoadmapService {
	return &RoadmapService{aiAdapter: aiAdapter}
}
