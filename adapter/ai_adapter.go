package adapter

import "context"

// AIAdapter wraps the Gemini API client.
type AIAdapter struct {
	// client *genai.Client  // TODO: add Gemini client
}

func NewAIAdapter( /* client *genai.Client */ ) *AIAdapter {
	return &AIAdapter{}
}

// Generate sends a prompt to the Gemini API and returns the response text.
func (a *AIAdapter) Generate(prompt string) (string, error) {
	_ = context.Background()
	// TODO: implement Gemini API call
	return "", nil
}
