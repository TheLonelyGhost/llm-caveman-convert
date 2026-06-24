package tokens

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

const googleAPIKeyEnv = "GOOGLE_API_KEY" //nolint:gosec // env var name, not a credential

const googleSkipReason = "requires GOOGLE_API_KEY"

type geminiCounter struct {
	model string
}

// newGeminiCounter returns a Counter for the given Gemini model.
// If GOOGLE_API_KEY is not set when Count is called, it returns a skip result.
func newGeminiCounter(model string) Counter {
	return &geminiCounter{model: model}
}

// Count calls the Gemini API CountTokens endpoint.
// Returns an error containing googleSkipReason when GOOGLE_API_KEY is absent.
func (c *geminiCounter) Count(ctx context.Context, text string) (int, error) {
	apiKey := os.Getenv(googleAPIKeyEnv)
	if apiKey == "" {
		return 0, fmt.Errorf("%s", googleSkipReason)
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{ //nolint:exhaustruct
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return 0, fmt.Errorf("gemini client: %w", err)
	}
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}
	resp, err := client.Models.CountTokens(ctx, c.model, contents, nil)
	if err != nil || resp == nil {
		if err == nil {
			err = fmt.Errorf("nil response")
		}
		return 0, fmt.Errorf("gemini count tokens: %w", err)
	}
	return int(resp.TotalTokens), nil
}
