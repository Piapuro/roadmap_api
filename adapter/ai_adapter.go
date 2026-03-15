package adapter

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

const geminiModel = "gemini-2.0-flash"

// AIAdapter wraps the Gemini API client.
type AIAdapter struct {
	client *genai.Client
}

func NewAIAdapter(client *genai.Client) *AIAdapter {
	return &AIAdapter{client: client}
}

// Generate sends a prompt to the Gemini API and returns the response text.
func (a *AIAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	result, err := a.client.Models.GenerateContent(ctx, geminiModel,
		genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}
	if result == nil || len(result.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned empty response")
	}
	return result.Text(), nil
}
