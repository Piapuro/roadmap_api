package service

import "github.com/Piapuro/roadmap_api/adapter"

// AIService handles business logic for AI-related features using Gemini.
type AIService struct {
	aiAdapter *adapter.AIAdapter
}

func NewAIService(aiAdapter *adapter.AIAdapter) *AIService {
	return &AIService{aiAdapter: aiAdapter}
}

// GenerateRoadmap sends a prompt to Gemini and returns the generated roadmap content.
func (s *AIService) GenerateRoadmap(prompt string) (string, error) {
	// TODO: implement
	return s.aiAdapter.Generate(prompt)
}
