package driver

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

func NewGeminiClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return client, nil
}
